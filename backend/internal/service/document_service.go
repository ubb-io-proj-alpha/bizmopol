package service

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"backend/internal/dto"
	"backend/internal/model"
	"backend/internal/repository"
	"backend/internal/ws"

	"github.com/google/uuid"
	pdfapi "github.com/pdfcpu/pdfcpu/pkg/api"
	pdfmodel "github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

const (
	uploadDir    = "/app/data/documents"
	signatureDir = "/app/data/signatures"
	maxFileSize  = 50 << 20
)

type DocumentService interface {
	Upload(ctx context.Context, userID string, req dto.DocumentUploadRequest, file multipart.File, header *multipart.FileHeader) (*model.Document, error)
	Get(ctx context.Context, id string) (*model.Document, error)
	List(ctx context.Context, q dto.DocumentQuery) ([]*model.Document, int64, error)
	Delete(ctx context.Context, id, userID string) error
	GetFilePath(ctx context.Context, id, mode string) (string, string, error)
	Sign(ctx context.Context, docID, userID, userEmail, ip string, req dto.DocumentSignRequest) (*model.Signature, error)
	GetSignatureImage(ctx context.Context, sigID string) (string, error)
	GetAuditLog(ctx context.Context, docID string) ([]*model.SignatureLog, error)
}

type documentService struct {
	repo        repository.DocumentRepository
	contactRepo repository.ContactRepository
	hub         *ws.Hub
}

func NewDocumentService(repo repository.DocumentRepository, contactRepo repository.ContactRepository, hub *ws.Hub) DocumentService {
	os.MkdirAll(uploadDir, 0755)
	os.MkdirAll(signatureDir, 0755)
	return &documentService{repo: repo, contactRepo: contactRepo, hub: hub}
}

func (s *documentService) Upload(ctx context.Context, userID string, req dto.DocumentUploadRequest, file multipart.File, header *multipart.FileHeader) (*model.Document, error) {
	if header.Size > maxFileSize {
		return nil, fmt.Errorf("file too large (max 50MB)")
	}

	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext != ".pdf" {
		return nil, fmt.Errorf("only PDF files are allowed")
	}

	id := uuid.New().String()
	originalName := id + "_original" + ext
	originalPath := filepath.Join(uploadDir, originalName)

	dst, err := os.Create(originalPath)
	if err != nil {
		return nil, fmt.Errorf("failed to save file")
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		os.Remove(originalPath)
		return nil, fmt.Errorf("failed to save file")
	}

	currentPath := filepath.Join(uploadDir, id+ext)
	copyFile(originalPath, currentPath)

	var contactName string
	if req.ContactID != "" {
		c, err := s.contactRepo.FindByID(ctx, req.ContactID)
		if err == nil && c != nil {
			contactName = c.Name
		}
	}

	doc := &model.Document{
		ID:           id,
		Title:        req.Title,
		FileName:     header.Filename,
		FilePath:     currentPath,
		OriginalPath: originalPath,
		FileSize:     header.Size,
		ContactID:    req.ContactID,
		ContactName:  contactName,
		UploadedBy:   userID,
		Status:       "pending",
	}

	if err := s.repo.CreateDocument(ctx, doc); err != nil {
		os.Remove(originalPath)
		os.Remove(currentPath)
		return nil, err
	}

	s.addLog(ctx, id, "uploaded", userID, "", "", fmt.Sprintf("Przesłano dokument: %s", header.Filename))

	return doc, nil
}

func (s *documentService) Get(ctx context.Context, id string) (*model.Document, error) {
	return s.repo.GetDocument(ctx, id)
}

func (s *documentService) List(ctx context.Context, q dto.DocumentQuery) ([]*model.Document, int64, error) {
	return s.repo.ListDocuments(ctx, q.Search, q.Status, q.ContactID, q.Page, q.PageSize)
}

func (s *documentService) Delete(ctx context.Context, id, userID string) error {
	doc, err := s.repo.GetDocument(ctx, id)
	if err != nil {
		return err
	}
	if doc == nil {
		return fmt.Errorf("not found")
	}

	os.Remove(doc.FilePath)
	os.Remove(doc.OriginalPath)
	os.Remove(doc.SignedFilePath)
	for _, sig := range doc.Signatures {
		os.Remove(sig.ImagePath)
	}

	return s.repo.DeleteDocument(ctx, id)
}

func (s *documentService) GetFilePath(ctx context.Context, id, mode string) (string, string, error) {
	doc, err := s.repo.GetDocument(ctx, id)
	if err != nil {
		return "", "", err
	}
	if doc == nil {
		return "", "", fmt.Errorf("not found")
	}

	switch mode {
	case "original":
		if doc.OriginalPath != "" {
			return doc.OriginalPath, doc.FileName, nil
		}
		return doc.FilePath, doc.FileName, nil
	default:
		if doc.SignedFilePath != "" {
			return doc.SignedFilePath, doc.FileName, nil
		}
		return doc.FilePath, doc.FileName, nil
	}
}

func (s *documentService) Sign(ctx context.Context, docID, userID, userEmail, ip string, req dto.DocumentSignRequest) (*model.Signature, error) {
	doc, err := s.repo.GetDocument(ctx, docID)
	if err != nil {
		return nil, err
	}
	if doc == nil {
		return nil, fmt.Errorf("not found")
	}

	imgData, err := base64.StdEncoding.DecodeString(req.ImageData)
	if err != nil {
		return nil, fmt.Errorf("invalid signature image data")
	}

	sigID := uuid.New().String()
	imgPath := filepath.Join(signatureDir, sigID+".png")

	if err := os.WriteFile(imgPath, imgData, 0644); err != nil {
		return nil, fmt.Errorf("failed to save signature")
	}

	sig := &model.Signature{
		ID:          sigID,
		DocumentID:  docID,
		SignerName:  req.SignerName,
		SignerEmail: req.SignerEmail,
		SignerID:    userID,
		ImagePath:   imgPath,
		IPAddress:   ip,
	}

	if err := s.repo.CreateSignature(ctx, sig); err != nil {
		os.Remove(imgPath)
		return nil, err
	}

	signedPath, err := s.stampSignatureOnPDF(doc, sig, req.SignerName)
	if err != nil {
		slog.Error("Failed to stamp signature on PDF", "error", err)
	} else {
		doc.SignedFilePath = signedPath
	}

	now := time.Now()
	doc.Status = "signed"
	doc.SignedAt = &now
	s.repo.UpdateDocument(ctx, doc)

	s.addLog(ctx, docID, "signed", userID, req.SignerName, ip, fmt.Sprintf("Podpisano przez: %s (%s)", req.SignerName, req.SignerEmail))

	if s.hub != nil {
		s.hub.SendToUser(doc.UploadedBy, ws.Notification{
			Type:    "document_signed",
			Message: fmt.Sprintf("Dokument \"%s\" został podpisany przez %s", doc.Title, req.SignerName),
			Data: map[string]interface{}{
				"document_id":    docID,
				"document_title": doc.Title,
				"signer_name":    req.SignerName,
			},
		})
	}

	return sig, nil
}

func (s *documentService) stampSignatureOnPDF(doc *model.Document, sig *model.Signature, signerName string) (string, error) {
	srcPath := doc.FilePath
	if doc.SignedFilePath != "" {
		srcPath = doc.SignedFilePath
	}

	signedPath := filepath.Join(uploadDir, doc.ID+"_signed.pdf")
	tmpPath := signedPath + ".tmp"

	conf := pdfmodel.NewDefaultConfiguration()

	sigCount, _ := s.repo.ListSignaturesByDocument(context.Background(), doc.ID)
	sigIndex := len(sigCount)
	yOffset := 30 + (sigIndex-1)*80

	desc := fmt.Sprintf("position:bl, scalefactor:0.25 abs, opacity:1.0, offset:50 %d", yOffset)
	err := pdfapi.AddImageWatermarksFile(srcPath, tmpPath, nil, true, sig.ImagePath, desc, conf)
	if err != nil {
		os.Remove(tmpPath)
		return "", fmt.Errorf("stamp image: %w", err)
	}

	labelDesc := fmt.Sprintf("position:bl, scalefactor:1.0 abs, opacity:1.0, offset:200 %d, points:8, color:0.2 0.2 0.2", yOffset+10)
	labelText := fmt.Sprintf("%s  |  %s", signerName, time.Now().Format("2006-01-02 15:04"))
	tmpPath2 := signedPath + ".tmp2"
	err = pdfapi.AddTextWatermarksFile(tmpPath, tmpPath2, nil, true, labelText, labelDesc, conf)
	if err != nil {
		os.Remove(tmpPath)
		os.Remove(tmpPath2)
		return "", fmt.Errorf("stamp text: %w", err)
	}

	os.Remove(tmpPath)
	os.Rename(tmpPath2, signedPath)

	return signedPath, nil
}

func (s *documentService) GetSignatureImage(ctx context.Context, sigID string) (string, error) {
	sig, err := s.repo.GetSignature(ctx, sigID)
	if err != nil {
		return "", err
	}
	if sig == nil {
		return "", fmt.Errorf("not found")
	}
	return sig.ImagePath, nil
}

func (s *documentService) GetAuditLog(ctx context.Context, docID string) ([]*model.SignatureLog, error) {
	return s.repo.ListLogsByDocument(ctx, docID)
}

func (s *documentService) addLog(ctx context.Context, docID, action, actorID, actorName, ip, details string) {
	log := &model.SignatureLog{
		ID:         uuid.New().String(),
		DocumentID: docID,
		Action:     action,
		ActorName:  actorName,
		ActorID:    actorID,
		IPAddress:  ip,
		Details:    details,
	}
	if err := s.repo.CreateLog(ctx, log); err != nil {
		slog.Error("Failed to create audit log", "error", err)
	}
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}
