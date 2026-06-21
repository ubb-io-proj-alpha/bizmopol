package dto

type DocumentUploadRequest struct {
	Title     string `form:"title" binding:"required"`
	ContactID string `form:"contact_id"`
}

type DocumentSignRequest struct {
	SignerName  string `json:"signer_name" binding:"required"`
	SignerEmail string `json:"signer_email"`
	ImageData   string `json:"image_data" binding:"required"`
}

type DocumentQuery struct {
	Search    string `form:"search"`
	Status    string `form:"status"`
	ContactID string `form:"contact_id"`
	Page      int    `form:"page"`
	PageSize  int    `form:"page_size"`
}
