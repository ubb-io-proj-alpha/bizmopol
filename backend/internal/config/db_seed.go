package config

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"backend/internal/model"
)

func SeedDatabase(db *gorm.DB) {
	seedUsers(db)
	seedContacts(db)
}

func seedUsers(db *gorm.DB) {
	users := []struct {
		email    string
		password string
		name     string
		role     model.Role
	}{
		{"admin@example.com", "password123", "Alice Smith", model.RoleAdmin},
		{"coach@example.com", "password123", "Bob Jones", model.RoleCoach},
		{"user@example.com", "password123", "Carol White", model.RoleUser},
	}

	for _, u := range users {
		var existing model.User
		if err := db.Where("email = ?", u.email).First(&existing).Error; err == nil {
			continue
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(u.password), bcrypt.DefaultCost)
		if err != nil {
			slog.Error("SeedDatabase: failed to hash password", "email", u.email, "error", err)
			continue
		}
		user := model.User{
			ID:           uuid.NewString(),
			Email:        u.email,
			PasswordHash: string(hash),
			Name:         u.name,
			Role:         u.role,
		}
		if err := db.Create(&user).Error; err != nil {
			slog.Error("SeedDatabase: failed to create user", "email", u.email, "error", err)
		}
	}
}

func seedContacts(db *gorm.DB) {
	var count int64
	db.Model(&model.Contact{}).Count(&count)
	if count > 0 {
		return
	}

	slog.Info("Seeding contacts...")

	adminID := getAdminID(db)

	now := time.Now()

	singleContacts := []struct {
		name    string
		email   string
		phone   string
		company string
		status  string
		notes   string
		daysAgo int
	}{
		{"Jan Kowalski", "jan.kowalski@techpol.pl", "+48 501 234 567", "TechPol", "lead", "Zainteresowany ofertą premium. Kontakt z polecenia Piotra.", 120},
		{"Karolina Nowak", "k.nowak@innosoft.pl", "+48 502 345 678", "InnoSoft", "prospect", "Po demo produktu, czeka na ofertę. Preferuje kontakt mailowy.", 90},
		{"Piotr Wiśniewski", "p.wisniewski@dataflow.pl", "+48 503 456 789", "DataFlow", "customer", "Aktywny klient od 2 lat. Używa pakietu Enterprise.", 60},
		{"Maria Wójcik", "m.wojcik@cloudbase.pl", "+48 504 567 890", "CloudBase", "inactive", "Nie odpowiada od 3 miesięcy. Kontrakt wygasł.", 200},
		{"Krzysztof Kamiński", "k.kaminski@netsolutions.pl", "+48 505 678 901", "NetSolutions", "lead", "Pierwszy kontakt przez formularz na stronie. Do zakwalifikowania.", 10},
		{"Katarzyna Lewandowska", "kat.lewandowska@digimark.pl", "+48 506 789 012", "DigiMark", "prospect", "Uczestniczka webinaru. Zainteresowana modułem CRM.", 45},
		{"Michał Zieliński", "m.zielinski@webcraft.pl", "+48 507 890 123", "WebCraft", "customer", "Klient od roku. Płaci regularnie. Polecił 2 nowych klientów.", 300},
		{"Honorata Szymańska", "h.szymanska@softhouse.pl", "+48 508 901 234", "SoftHouse", "inactive", "Zrezygnowała po 6 miesiącach. Powód: zmiana dostawcy.", 150},
		{"Tomasz Woźniak", "t.wozniak@codelab.pl", "+48 509 012 345", "CodeLab", "lead", "Pobrał e-book. Nie odpowiedział na follow-up.", 5},
		{"Szymon Kowalczyk", "sz.kowalczyk@pixelstudio.pl", "+48 510 123 456", "PixelStudio", "prospect", "Negocjacje w toku. Interesuje go pakiet startowy.", 30},
		{"Anna Dąbrowska", "a.dabrowska@innovit.pl", "+48 511 234 567", "InnoVit", "customer", "VIP. Płaci za plan roczny. Bardzo zadowolona z wsparcia.", 400},
		{"Marcin Jabłoński", "m.jablonski@bizpro.pl", "+48 512 345 678", "BizPro", "lead", "Zimny lead z kampanii LinkedIn.", 3},
		{"Agnieszka Michalska", "a.michalska@ecomplus.pl", "+48 513 456 789", "EcomPlus", "prospect", "Porównuje nas z konkurencją. Ma pytania o API.", 20},
		{"Łukasz Pawlak", "l.pawlak@startex.pl", "+48 514 567 890", "StartEx", "customer", "Małe przedsiębiorstwo. Używa planu Basic.", 180},
		{"Natalia Krawczyk", "n.krawczyk@growfast.pl", "+48 515 678 901", "GrowFast", "inactive", "Nie przedłużył subskrypcji. Wyjechał za granicę.", 250},
	}

	contactIDs := make(map[string]string)
	for _, sc := range singleContacts {
		id := uuid.NewString()
		contactIDs[sc.name] = id
		createdAt := now.AddDate(0, 0, -sc.daysAgo)
		c := model.Contact{
			ID:        id,
			Name:      sc.name,
			Email:     sc.email,
			Phone:     sc.phone,
			Company:   sc.company,
			Status:    sc.status,
			Notes:     sc.notes,
			CreatedAt: createdAt,
			UpdatedAt: createdAt,
		}
		if err := db.Create(&c).Error; err != nil {
			slog.Error("SeedDatabase: failed to create contact", "name", sc.name, "error", err)
			continue
		}
		seedContactHistory(db, id, sc.status, sc.daysAgo, adminID)
	}

	mergeGroups := []struct {
		groupName string
		company   string
		status    string
		members   []string
	}{
		{
			groupName: "TechPol Group",
			company:   "TechPol",
			status:    "customer",
			members:   []string{"Jan Kowalski", "Karolina Nowak"},
		},
		{
			groupName: "DataFlow & CloudBase Alliance",
			company:   "DataFlow",
			status:    "prospect",
			members:   []string{"Piotr Wiśniewski", "Maria Wójcik", "Krzysztof Kamiński"},
		},
	}

	for _, mg := range mergeGroups {
		groupID := uuid.NewString()
		group := model.Contact{
			ID:        groupID,
			Name:      mg.groupName,
			Company:   mg.company,
			Status:    mg.status,
			IsGroup:   true,
			CreatedAt: now.AddDate(0, 0, -15),
			UpdatedAt: now.AddDate(0, 0, -15),
		}
		if err := db.Create(&group).Error; err != nil {
			slog.Error("SeedDatabase: failed to create group", "name", mg.groupName, "error", err)
			continue
		}
		for _, memberName := range mg.members {
			memberID, ok := contactIDs[memberName]
			if !ok {
				continue
			}
			db.Model(&model.Contact{}).Where("id = ?", memberID).Updates(map[string]interface{}{
				"group_id": groupID,
			})
		}
		db.Create(&model.ContactHistory{
			ID:          uuid.NewString(),
			ContactID:   groupID,
			Action:      "merged",
			Description: fmt.Sprintf("Połączono %d kontaktów w grupę: %s", len(mg.members), mg.groupName),
			UserID:      adminID,
			CreatedAt:   now.AddDate(0, 0, -15),
		})
		db.Create(&model.ContactHistory{
			ID:          uuid.NewString(),
			ContactID:   groupID,
			Action:      "note",
			Description: "Inicjalna konfiguracja grupy zakończona. Przypisano opiekuna klienta.",
			UserID:      adminID,
			CreatedAt:   now.AddDate(0, 0, -14),
		})
	}

	extraContacts := []struct {
		name    string
		email   string
		phone   string
		company string
		status  string
		notes   string
	}{
		{"Dominika Kowal", "d.kowal@mediapro.pl", "+48 601 111 001", "MediaPro", "lead", ""},
		{"Robert Nowicki", "r.nowicki@salesforce.pl", "+48 601 111 002", "SalesForce PL", "prospect", ""},
		{"Beata Wiśniewska", "b.wisniewska@hr-plus.pl", "+48 601 111 003", "HR Plus", "customer", ""},
		{"Grzegorz Wójcicki", "g.wojcicki@fintech.pl", "+48 601 111 004", "FinTech", "inactive", ""},
		{"Marta Kamińska", "m.kaminska@retail.pl", "+48 601 111 005", "Retail Inc", "lead", ""},
		{"Paweł Lewandowski", "p.lewandowski@b2b.pl", "+48 601 111 006", "B2B Corp", "prospect", ""},
		{"Sylwia Zielińska", "s.zielinska@agency.pl", "+48 601 111 007", "Agency One", "customer", ""},
		{"Rafał Szymański", "r.szymanski@logistic.pl", "+48 601 111 008", "Logistic Hub", "lead", ""},
		{"Joanna Woźniacka", "j.wozniacka@healthco.pl", "+48 601 111 009", "HealthCo", "prospect", ""},
		{"Bartosz Kowalczewski", "b.kowalczewski@studio.pl", "+48 601 111 010", "Studio Digital", "customer", ""},
		{"Ewa Dąbrowska", "e.dabrowska@edutech.pl", "+48 601 111 011", "EduTech", "lead", ""},
		{"Marek Jabłoński", "m.jablonski2@proptech.pl", "+48 601 111 012", "PropTech", "inactive", ""},
		{"Monika Michalska", "mo.michalska@green.pl", "+48 601 111 013", "GreenSolutions", "prospect", ""},
		{"Dariusz Pawlak", "d.pawlak@transport.pl", "+48 601 111 014", "TransPol", "customer", ""},
		{"Izabela Krawczyk", "i.krawczyk@fashion.pl", "+48 601 111 015", "FashionHub", "lead", ""},
		{"Wojciech Kowalski", "w.kowalski2@autotech.pl", "+48 601 111 016", "AutoTech", "prospect", ""},
		{"Małgorzata Nowak", "ma.nowak@consulting.pl", "+48 601 111 017", "Consulting Plus", "customer", ""},
		{"Tomasz Wiśniewski", "to.wisniewski@realty.pl", "+48 601 111 018", "Realty Group", "inactive", ""},
		{"Kamila Wójcik", "ka.wojcik@startup.pl", "+48 601 111 019", "StartupZone", "lead", ""},
		{"Sebastian Kamiński", "se.kaminski@cloud9.pl", "+48 601 111 020", "Cloud9", "prospect", ""},
		{"Agata Lewandowska", "ag.lewandowska@shop.pl", "+48 601 111 021", "ShopPro", "customer", ""},
		{"Damian Zieliński", "da.zielinski@gaming.pl", "+48 601 111 022", "GameStudio", "lead", ""},
		{"Patrycja Szymańska", "pa.szymanska@media.pl", "+48 601 111 023", "MediaGroup", "prospect", ""},
		{"Konrad Woźniak", "ko.wozniak@infra.pl", "+48 601 111 024", "InfraNet", "customer", ""},
		{"Justyna Kowalczyk", "ju.kowalczyk@bio.pl", "+48 601 111 025", "BioPharma", "inactive", ""},
		{"Arkadiusz Dąbrowski", "ar.dabrowski@market.pl", "+48 601 111 026", "MarketPro", "lead", ""},
		{"Weronika Jabłońska", "we.jablonska@legal.pl", "+48 601 111 027", "LegalDesk", "prospect", ""},
		{"Przemysław Michalski", "pr.michalski@it.pl", "+48 601 111 028", "ITSolutions", "customer", ""},
		{"Aleksandra Pawlak", "al.pawlak@hr.pl", "+48 601 111 029", "HR Connect", "lead", ""},
		{"Krystian Krawczyk", "kr.krawczyk@payment.pl", "+48 601 111 030", "PaySmart", "prospect", ""},
		{"Żaneta Kowal", "za.kowal@ads.pl", "+48 601 111 031", "AdsAgency", "customer", ""},
		{"Łukasz Nowicki", "lu.nowicki@web.pl", "+48 601 111 032", "WebFactory", "inactive", ""},
		{"Natalia Wiśniewska", "na.wisniewska@food.pl", "+48 601 111 033", "FoodTech", "lead", ""},
		{"Adrian Wójcicki", "ad.wojcicki@clean.pl", "+48 601 111 034", "CleanPro", "prospect", ""},
		{"Karol Kamińska", "ka2.kaminska@build.pl", "+48 601 111 035", "BuildCorp", "customer", ""},
	}

	for i, sc := range extraContacts {
		id := uuid.NewString()
		daysAgo := 5 + i*3
		createdAt := now.AddDate(0, 0, -daysAgo)
		c := model.Contact{
			ID:        id,
			Name:      sc.name,
			Email:     sc.email,
			Phone:     sc.phone,
			Company:   sc.company,
			Status:    sc.status,
			Notes:     sc.notes,
			CreatedAt: createdAt,
			UpdatedAt: createdAt,
		}
		if err := db.Create(&c).Error; err != nil {
			slog.Error("SeedDatabase: failed to create extra contact", "name", sc.name, "error", err)
			continue
		}
		db.Create(&model.ContactHistory{
			ID:          uuid.NewString(),
			ContactID:   id,
			Action:      "created",
			Description: "Kontakt został dodany do systemu.",
			UserID:      adminID,
			CreatedAt:   createdAt,
		})
	}
}

func seedContactHistory(db *gorm.DB, contactID, status string, daysAgo int, userID string) {
	now := time.Now()
	base := now.AddDate(0, 0, -daysAgo)

	events := []struct {
		action      string
		description string
		offsetDays  int
	}{
		{"created", "Kontakt został dodany do systemu CRM.", 0},
	}

	switch status {
	case "prospect":
		events = append(events,
			struct {
				action      string
				description string
				offsetDays  int
			}{"call", "Pierwsza rozmowa telefoniczna. Klient zainteresowany ofertą.", 2},
			struct {
				action      string
				description string
				offsetDays  int
			}{"email", "Wysłano prezentację produktu i cennik.", 4},
			struct {
				action      string
				description string
				offsetDays  int
			}{"note", "Klient porównuje oferty. Decyzja spodziewana w ciągu 2 tygodni.", 7},
		)
	case "customer":
		events = append(events,
			struct {
				action      string
				description string
				offsetDays  int
			}{"call", "Rozmowa kwalifikacyjna. Pozytywna reakcja na ofertę.", 3},
			struct {
				action      string
				description string
				offsetDays  int
			}{"email", "Wysłano umowę do podpisania.", 7},
			struct {
				action      string
				description string
				offsetDays  int
			}{"updated", "Status zmieniony z prospect na customer po podpisaniu umowy.", 10},
			struct {
				action      string
				description string
				offsetDays  int
			}{"meeting", "Spotkanie onboardingowe. Omówiono wdrożenie i szkolenie zespołu.", 15},
			struct {
				action      string
				description string
				offsetDays  int
			}{"note", "Wdrożenie zakończone. Klient aktywnie używa platformy.", 20},
			struct {
				action      string
				description string
				offsetDays  int
			}{"call", "Miesięczna rozmowa przeglądowa. Klient zadowolony z wyników.", 30},
		)
	case "inactive":
		events = append(events,
			struct {
				action      string
				description string
				offsetDays  int
			}{"call", "Próba kontaktu — brak odpowiedzi.", 5},
			struct {
				action      string
				description string
				offsetDays  int
			}{"email", "Wysłano e-mail z pytaniem o zainteresowanie odnowieniem.", 10},
			struct {
				action      string
				description string
				offsetDays  int
			}{"note", "Klient nie odpowiada. Prawdopodobnie zmienił dostawcę.", 20},
			struct {
				action      string
				description string
				offsetDays  int
			}{"updated", "Status zmieniony na inactive po braku odpowiedzi przez 30 dni.", 30},
		)
	case "lead":
		events = append(events,
			struct {
				action      string
				description string
				offsetDays  int
			}{"note", "Lead z kampanii marketingowej. Wymaga kwalifikacji.", 1},
		)
	}

	for _, e := range events {
		t := base.AddDate(0, 0, e.offsetDays)
		if t.After(now) {
			t = now
		}
		db.Create(&model.ContactHistory{
			ID:          uuid.NewString(),
			ContactID:   contactID,
			Action:      e.action,
			Description: e.description,
			UserID:      userID,
			CreatedAt:   t,
		})
	}
}

func getAdminID(db *gorm.DB) string {
	var user model.User
	if err := db.Where("email = ?", "admin@example.com").First(&user).Error; err != nil {
		return fmt.Sprintf("system-%s", uuid.NewString()[:8])
	}
	return user.ID
}
