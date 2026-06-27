package config

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"backend/internal/model"
	"backend/internal/service"
)

func SeedDatabase(db *gorm.DB) {
	seedUsers(db)
	seedContacts(db)
	seedPipelines(db)
	seedCommunication(db)
	seedFunnels(db)
	seedBooking(db)
}

func seedBooking(db *gorm.DB) {
	var count int64
	db.Model(&model.BookingSettings{}).Count(&count)
	if count > 0 {
		return
	}

	var admin model.User
	if err := db.Where("email = ?", "admin@example.com").First(&admin).Error; err != nil {
		slog.Error("SeedDatabase: booking owner (admin) not found", "error", err)
		return
	}

	slog.Info("Seeding booking settings...")
	settings := model.BookingSettings{
		ID:           model.BookingSettingsID,
		OwnerUserID:  admin.ID,
		Enabled:      true,
		WorkingDays:  "1,2,3,4,5",
		StartMinutes: 540,
		EndMinutes:   1020,
		SlotMinutes:  30,
		MeetingTitle: "Spotkanie: {name}",
	}
	if err := db.Create(&settings).Error; err != nil {
		slog.Error("SeedDatabase: failed to create booking settings", "error", err)
	}
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

func calcSeedScore(db *gorm.DB, contactID string, c *model.Contact) int {
	var count int64
	db.Model(&model.ContactHistory{}).Where("contact_id = ?", contactID).Count(&count)

	lastActivityDays := -1
	var last model.ContactHistory
	if err := db.Where("contact_id = ?", contactID).Order("created_at DESC").First(&last).Error; err == nil {
		lastActivityDays = service.DaysSince(last.CreatedAt)
	}

	return service.CalcLeadScore(c, int(count), lastActivityDays)
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
		c.LeadScore = calcSeedScore(db, id, &c)
		db.Model(&model.Contact{}).Where("id = ?", id).Update("lead_score", c.LeadScore)
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
		group.LeadScore = calcSeedScore(db, groupID, &group)
		db.Model(&model.Contact{}).Where("id = ?", groupID).Update("lead_score", group.LeadScore)
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
		c.LeadScore = calcSeedScore(db, id, &c)
		db.Model(&model.Contact{}).Where("id = ?", id).Update("lead_score", c.LeadScore)
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
			}{"call", "Próba kontaktu - brak odpowiedzi.", 5},
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

func seedFunnels(db *gorm.DB) {
    var count int64
    db.Model(&model.Funnel{}).Count(&count)
    if count > 0 {
        return
    }

    slog.Info("Seeding funnels and pages...")

    funnelsData := []struct {
        name         string
        subdomain    string
        customDomain string
        pages        []struct {
            name string
            path string
            html string
            css  string
        }
    }{
        {
            name:      "Kampania Wiosenna 2026 - Lead Magnet",
            subdomain: "wiosna2026.bizmopol.localhost",
            pages: []struct {
                name, path, html, css string
            }{
                {
                    name: "Strona Zapisu (Opt-in)",
                    path: "/",
                    html: `<div class="container"><h1>Pobierz darmowy E-book!</h1><p>Zostaw maila, aby otrzymać poradnik.</p><button>Zapisz się</button></div>`,
                    css:  `.container { text-align: center; font-family: sans-serif; padding: 50px; } button { background: #007bff; color: white; padding: 10px 20px; border: none; border-radius: 5px; }`,
                },
                {
                    name: "Strona Podziękowania (Thank you)",
                    path: "/sukces",
                    html: `<div class="container"><h1>Dziękujemy!</h1><p>E-book leci na Twoją skrzynkę.</p></div>`,
                    css:  `.container { text-align: center; font-family: sans-serif; padding: 50px; color: green; }`,
                },
            },
        },
        {
            name:         "Webinar B2B - Sprzedażowy",
            subdomain:    "webinar-b2b",
            customDomain: "www.moj-super-webinar.pl",
            pages: []struct {
                name, path, html, css string
            }{
                {
                    name: "Rejestracja",
                    path: "/",
                    html: `<div class="hero"><h1>Zbuduj system CRM w 30 dni</h1><p>Webinar na żywo - wtorek, 18:00</p></div>`,
                    css:  `.hero { background-color: #1a1a1a; color: #ffffff; padding: 100px 20px; text-align: center; }`,
                },
            },
        },
        {
            name:      "Krótki Link - Bio Instagram",
            subdomain: "linki.bizmopol.localhost",
            pages: []struct {
                name, path, html, css string
            }{
                {
                    name: "Linktree Clone",
                    path: "/jan-kowalski",
                    html: `<div class="links"><a href="#">Mój Blog</a><a href="#">Mój Sklep</a><a href="#">Konsultacje</a></div>`,
                    css:  `.links { display: flex; flex-direction: column; gap: 15px; max-width: 400px; margin: 40px auto; } .links a { display: block; padding: 15px; background: #eee; text-decoration: none; color: #333; text-align: center; border-radius: 8px; }`,
                },
            },
        },
    }

    for _, fData := range funnelsData {
        funnelID := uuid.NewString()
        f := model.Funnel{
            ID:           funnelID,
            Name:         fData.name,
            Subdomain:    fData.subdomain,
            CustomDomain: fData.customDomain,
        }

        if err := db.Create(&f).Error; err != nil {
            slog.Error("SeedDatabase: failed to create funnel", "name", f.Name, "error", err)
            continue
        }

        for _, pData := range fData.pages {
            p := model.Page{
                ID:          uuid.NewString(),
                FunnelID:    funnelID,
                Name:        pData.name,
                Path:        pData.path,
                Structure:   `{"blocks": []}`,
                HTMLContent: pData.html,
                CSSContent:  pData.css,
            }

            if err := db.Create(&p).Error; err != nil {
                slog.Error("SeedDatabase: failed to create page", "funnel", f.Name, "page", p.Name, "error", err)
            }
        }
    }
}

func getAdminID(db *gorm.DB) string {
	var user model.User
	if err := db.Where("email = ?", "admin@example.com").First(&user).Error; err != nil {
		return fmt.Sprintf("system-%s", uuid.NewString()[:8])
	}
	return user.ID
}

func seedPipelines(db *gorm.DB) {
	var count int64
	db.Model(&model.Pipeline{}).Count(&count)
	slog.Info("Seeding pipelines...")

	adminID := getAdminID(db)

	pipelines := []struct {
		name        string
		description string
		stages      []struct {
			name      string
			color     string
			sortOrder int
		}
	}{
		{
			name:        "Sprzedaż B2B",
			description: "Główny proces sprzedaży dla klientów biznesowych",
			stages: []struct {
				name      string
				color     string
				sortOrder int
			}{
				{"Nowy Lead", "#38bdf8", 0},
				{"Kwalifikacja", "#818cf8", 1},
				{"Demo / Prezentacja", "#f59e0b", 2},
				{"Oferta wysłana", "#fb923c", 3},
				{"Negocjacje", "#f87171", 4},
				{"Zamknięty – Wygrany", "#22c55e", 5},
				{"Zamknięty – Przegrany", "#64748b", 6},
			},
		},
		{
			name:        "Onboarding Klientów",
			description: "Proces wdrożenia nowych klientów po podpisaniu umowy",
			stages: []struct {
				name      string
				color     string
				sortOrder int
			}{
				{"Podpisana umowa", "#38bdf8", 0},
				{"Konfiguracja konta", "#818cf8", 1},
				{"Szkolenie zespołu", "#f59e0b", 2},
				{"Pierwsze wdrożenie", "#fb923c", 3},
				{"Aktywny klient", "#22c55e", 4},
			},
		},
		{
			name:        "Obsługa Leadów Marketingowych",
			description: "Przetwarzanie leadów przychodzących z kampanii",
			stages: []struct {
				name      string
				color     string
				sortOrder int
			}{
				{"Nowy kontakt", "#38bdf8", 0},
				{"Kontakt nawiązany", "#818cf8", 1},
				{"Zainteresowany", "#f59e0b", 2},
				{"Przekazany do sprzedaży", "#22c55e", 3},
				{"Odrzucony", "#64748b", 4},
            },
        },
    }

	for _, pd := range pipelines {
		pipelineID := uuid.NewString()
		now := time.Now()
		pipeline := model.Pipeline{
			ID:          pipelineID,
			Name:        pd.name,
			Description: pd.description,
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		stages := make([]model.Stage, 0, len(pd.stages))
		for _, sd := range pd.stages {
			stages = append(stages, model.Stage{
				ID:         uuid.NewString(),
				PipelineID: pipelineID,
				Name:       sd.name,
				Color:      sd.color,
				SortOrder:  sd.sortOrder,
				CreatedAt:  now,
				UpdatedAt:  now,
			})
		}
		pipeline.Stages = stages

		if err := db.Create(&pipeline).Error; err != nil {
			slog.Error("SeedDatabase: failed to create pipeline", "name", pd.name, "error", err)
			continue
		}

		seedContactStages(db, pipelineID, stages, adminID)
	}
}

func seedContactStages(db *gorm.DB, pipelineID string, stages []model.Stage, adminID string) {
	if len(stages) == 0 {
		return
	}

	var contacts []*model.Contact
	db.Where("is_group = ? AND group_id = ?", false, "").Limit(20).Find(&contacts)
	if len(contacts) == 0 {
		return
	}

	now := time.Now()
	contactsPerStage := max(1, len(contacts)/len(stages))

	for i, contact := range contacts {
		stageIdx := i / contactsPerStage
		if stageIdx >= len(stages) {
			stageIdx = len(stages) - 1
		}
		stage := stages[stageIdx]

		cs := model.ContactStage{
			ID:         uuid.NewString(),
			ContactID:  contact.ID,
			PipelineID: pipelineID,
			StageID:    stage.ID,
			CreatedAt:  now,
			UpdatedAt:  now,
		}
		if err := db.Create(&cs).Error; err != nil {
			slog.Error("SeedDatabase: failed to create contact stage", "contact", contact.ID, "error", err)
			continue
		}

		db.Create(&model.ContactHistory{
			ID:          uuid.NewString(),
			ContactID:   contact.ID,
			Action:      "stage_change",
			Description: fmt.Sprintf("Dodano do pipeline w etapie: %s", stage.Name),
			UserID:      adminID,
			CreatedAt:   now,
		})
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func seedCommunication(db *gorm.DB) {
	var count int64
	db.Model(&model.EmailThread{}).Count(&count)

	if count > 0 {
		return
	}

	slog.Info("Seeding communication data...")

	adminID := getAdminID(db)
	now := time.Now()

	templates := []model.EmailTemplate{
		{
			ID:       uuid.NewString(),
			Name:     "Powitanie nowego klienta",
			Subject:  "Witamy w BizmoPol!",
			Body:     "Dzień dobry {{name}},\n\nSerdecznie witamy Cię w gronie klientów BizmoPol.\n\nZ poważaniem,\nZespół BizmoPol",
			BodyHTML: "<p>Dzień dobry <strong>{{name}}</strong>,</p><p>Serdecznie witamy Cię w gronie klientów BizmoPol.</p><p>Z poważaniem,<br>Zespół BizmoPol</p>",
			Category: "Onboarding",
		},
		{
			ID:       uuid.NewString(),
			Name:     "Follow-up po demo",
			Subject:  "Dziękujemy za uczestnictwo w demo",
			Body:     "Dzień dobry {{name}},\n\nDziękujemy za udział w naszym demo. Chcielibyśmy poznać Twoje wrażenia.\n\nCzy masz pytania dotyczące naszego produktu?\n\nZ poważaniem,\nZespół BizmoPol",
			Category: "Sales",
		},
		{
			ID:       uuid.NewString(),
			Name:     "Reaktywacja nieaktywnego klienta",
			Subject:  "Tęsknimy za Tobą, {{name}}!",
			Body:     "Dzień dobry {{name}},\n\nZauważyliśmy, że od dłuższego czasu nie korzystałeś z naszej platformy. Chcielibyśmy zaproponować Ci specjalną ofertę powitalną.\n\nZ poważaniem,\nZespół BizmoPol",
			Category: "Retention",
		},
		{
			ID:       uuid.NewString(),
			Name:     "Oferta specjalna",
			Subject:  "Ekskluzywna oferta dla Ciebie!",
			Body:     "Dzień dobry {{name}},\n\nMamy dla Ciebie wyjątkową propozycję. Skontaktuj się z nami, aby dowiedzieć się więcej.\n\nZ poważaniem,\nZespół BizmoPol",
			Category: "Marketing",
		},
		{
			ID:       uuid.NewString(),
			Name:     "Przypomnienie o płatności",
			Subject:  "Przypomnienie o odnowieniu subskrypcji",
			Body:     "Dzień dobry {{name}},\n\nInformujemy, że Twoja subskrypcja wygasa wkrótce. Prosimy o jej odnowienie.\n\nZ poważaniem,\nZespół BizmoPol",
			Category: "Billing",
		},
	}

	for _, t := range templates {
		db.Create(&t)
	}

	var sig model.EmailSignature
	db.Where("user_id = ?", adminID).First(&sig)
	if sig.ID == "" {
		db.Create(&model.EmailSignature{
			ID:        uuid.NewString(),
			UserID:    adminID,
			Name:      "Podpis główny",
			Body:      "Z poważaniem,\nAlice Smith\nBizmoPol CRM\ntel: +48 100 200 300\nwww.bizmopol.pl",
			IsDefault: true,
		})
		db.Create(&model.EmailSignature{
			ID:     uuid.NewString(),
			UserID: adminID,
			Name:   "Podpis krótki",
			Body:   "Pozdrawiam,\nAlice | BizmoPol",
		})
	}

	contacts := []struct {
		name    string
		email   string
		company string
	}{
		{"Jan Kowalski", "jan.kowalski@techpol.pl", "TechPol"},
		{"Karolina Nowak", "k.nowak@innosoft.pl", "InnoSoft"},
		{"Piotr Wiśniewski", "p.wisniewski@dataflow.pl", "DataFlow"},
		{"Anna Dąbrowska", "a.dabrowska@innovit.pl", "InnoVit"},
		{"Michał Zieliński", "m.zielinski@webcraft.pl", "WebCraft"},
		{"Katarzyna Lewandowska", "kat.lewandowska@digimark.pl", "DigiMark"},
		{"Szymon Kowalczyk", "sz.kowalczyk@pixelstudio.pl", "PixelStudio"},
		{"Agnieszka Michalska", "a.michalska@ecomplus.pl", "EcomPlus"},
	}

	threadData := []struct {
		contactIdx int
		subject    string
		daysAgo    int
		msgs       []struct {
			dir  string
			body string
			days int
		}
	}{
		{
			contactIdx: 0,
			subject:    "Oferta na pakiet Enterprise",
			daysAgo:    30,
			msgs: []struct {
				dir  string
				body string
				days int
			}{
				{"outbound", "Dzień dobry Jan,\n\nPrzesyłam ofertę na pakiet Enterprise zgodnie z naszą rozmową.\n\nPozdrawiam,\nAlice", 30},
				{"inbound", "Dzień dobry Alice,\n\nDziękuję za ofertę. Przeanalizuję ją i wrócę do Pani do końca tygodnia.\n\nJan Kowalski", 28},
				{"outbound", "Dzień dobry Jan,\n\nCzy miał Pan czas zapoznać się z ofertą? Jestem do dyspozycji w przypadku pytań.\n\nPozdrawiam,\nAlice", 25},
				{"inbound", "Tak, oferta wygląda interesująco. Chciałbym omówić szczegóły implementacji.", 23},
			},
		},
		{
			contactIdx: 1,
			subject:    "Demo produktu - potwierdzenie",
			daysAgo:    15,
			msgs: []struct {
				dir  string
				body string
				days int
			}{
				{"outbound", "Szanowna Pani Karolino,\n\nPotwierdzam demo produktu na przyszły wtorek o godzinie 14:00.\n\nPozdrawiam,\nAlice", 15},
				{"inbound", "Dziękuję za potwierdzenie. Czy będzie można podczas demo zobaczyć moduł raportowania?\n\nPozdrawiam,\nKarolina", 14},
				{"outbound", "Oczywiście! Pokażemy cały moduł raportowania wraz z integracjami.\n\nDo zobaczenia we wtorek!\nAlice", 13},
			},
		},
		{
			contactIdx: 2,
			subject:    "Odnowienie kontraktu Enterprise",
			daysAgo:    7,
			msgs: []struct {
				dir  string
				body string
				days int
			}{
				{"inbound", "Dzień dobry,\n\nChciałbym omówić warunki odnowienia naszego kontraktu na kolejny rok.\n\nPiotr Wiśniewski\nDataFlow", 7},
				{"outbound", "Dzień dobry Piotrze,\n\nCieszę się, że chcecie z nami zostać! Przygotowuję specjalną ofertę dla lojalnych klientów.\n\nSkontaktuję się jutro.\nAlice", 6},
				{"inbound", "Świetnie! Czekam na ofertę. Zależy nam też na dodaniu 5 nowych użytkowników.", 5},
				{"outbound", "Oczywiście, uwzględnię to w ofercie. Przesyłam ją dziś po południu.\nAlice", 5},
				{"inbound", "Oferta wygląda bardzo dobrze. Potrzebuję tylko akceptacji zarządu.", 3},
			},
		},
		{
			contactIdx: 3,
			subject:    "Pytanie o API",
			daysAgo:    3,
			msgs: []struct {
				dir  string
				body string
				days int
			}{
				{"inbound", "Dzień dobry,\n\nMam pytanie dotyczące integracji z API. Czy jest dostępna dokumentacja?\n\nAnna Dąbrowska", 3},
				{"outbound", "Dzień dobry Anno,\n\nOczywiście! Dokumentację API znajdzie Pani pod adresem: docs.bizmopol.pl/api\n\nJeśli potrzebuje Pani pomocy, służę wsparciem.\nAlice", 2},
			},
		},
		{
			contactIdx: 4,
			subject:    "Newsletter - nowości produktowe Q1",
			daysAgo:    45,
			msgs: []struct {
				dir  string
				body string
				days int
			}{
				{"outbound", "Dzień dobry Michale,\n\nW tym kwartale wprowadziliśmy szereg nowości:\n- Nowy moduł raportowania\n- Integracja z Gmail\n- Ulepszone API\n\nSzczegóły na naszym blogu.\nPozdrawiam,\nAlice", 45},
			},
		},
		{
			contactIdx: 5,
			subject:    "Zaproszenie na webinar CRM",
			daysAgo:    20,
			msgs: []struct {
				dir  string
				body string
				days int
			}{
				{"outbound", "Szanowna Pani Katarzyno,\n\nZapraszamy na bezpłatny webinar 'Jak efektywnie zarządzać klientami w CRM'\n\nTermin: 15 maja, godz. 12:00\nLink: webinar.bizmopol.pl\n\nPozdrawiam,\nAlice", 20},
				{"inbound", "Dziękuję za zaproszenie! Już się zapisałam. Czy będzie nagranie po webinarze?\n\nKatarzyna", 19},
				{"outbound", "Tak, nagranie będzie dostępne przez 30 dni po webinarze.\n\nDo zobaczenia!\nAlice", 18},
			},
		},
	}

	for _, td := range threadData {
		if td.contactIdx >= len(contacts) {
			continue
		}
		contact := contacts[td.contactIdx]

		var dbContact model.Contact
		db.Where("email = ?", contact.email).First(&dbContact)
		contactID := dbContact.ID

		lastMsgTime := now.AddDate(0, 0, -td.daysAgo)
		if len(td.msgs) > 0 {
			lastMsgTime = now.AddDate(0, 0, -td.msgs[len(td.msgs)-1].days)
		}

		unread := 0
		for _, m := range td.msgs {
			if m.dir == "inbound" {
				unread++
			}
		}
		if unread > 0 {
			unread = 1
		}

		thread := model.EmailThread{
			ID:            uuid.NewString(),
			Subject:       td.subject,
			ContactID:     contactID,
			ContactName:   contact.name,
			ContactEmail:  contact.email,
			Status:        "open",
			Direction:     td.msgs[0].dir,
			LastMessageAt: lastMsgTime,
			MessageCount:  len(td.msgs),
			UnreadCount:   unread,
			CreatedAt:     now.AddDate(0, 0, -td.daysAgo),
		}
		db.Create(&thread)

		for _, m := range td.msgs {
			msgTime := now.AddDate(0, 0, -m.days)
			fromAddr := "crm@bizmopol.pl"
			toAddr := contact.email
			if m.dir == "inbound" {
				fromAddr = contact.email
				toAddr = "crm@bizmopol.pl"
			}
			isRead := m.dir == "outbound"
			sentAt := msgTime
			msg := model.EmailMessage{
				ID:        uuid.NewString(),
				ThreadID:  thread.ID,
				MessageID: uuid.NewString() + "@bizmopol",
				From:      fromAddr,
				To:        toAddr,
				Subject:   td.subject,
				Body:      m.body,
				Direction: m.dir,
				IsRead:    isRead,
				SentAt:    &sentAt,
				CreatedAt: msgTime,
			}
			db.Create(&msg)
		}
	}

	bulkJobID := uuid.NewString()
	completedAt := now.AddDate(0, 0, -14)
	startedAt := now.AddDate(0, 0, -14)
	db.Create(&model.BulkEmailJob{
		ID:          bulkJobID,
		Name:        "Kampania Q1 2025 - rejestracja webinaru",
		Subject:     "Zaproszenie na bezpłatny webinar!",
		Body:        "Zapraszamy na webinar CRM...",
		Status:      "completed",
		TotalCount:  50,
		SentCount:   48,
		FailedCount: 2,
		StartedAt:   &startedAt,
		CompletedAt: &completedAt,
		UserID:      adminID,
		CreatedAt:   now.AddDate(0, 0, -14),
	})

	bulkJobID2 := uuid.NewString()
	completedAt2 := now.AddDate(0, 0, -3)
	startedAt2 := now.AddDate(0, 0, -3)
	db.Create(&model.BulkEmailJob{
		ID:          bulkJobID2,
		Name:        "Reaktywacja klientów nieaktywnych",
		Subject:     "Tęsknimy za Tobą!",
		Body:        "Wróć do BizmoPol...",
		Status:      "completed",
		TotalCount:  25,
		SentCount:   24,
		FailedCount: 1,
		StartedAt:   &startedAt2,
		CompletedAt: &completedAt2,
		UserID:      adminID,
		CreatedAt:   now.AddDate(0, 0, -3),
	})

	slog.Info("Communication seeding completed")
}
