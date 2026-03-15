package config

import (
	"fmt"
    "log/slog"

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

    firstNames := []string{"Jan", "Karolina", "Piotr", "Maria", "Krzysztof", "Katarzyna", "Michał", "Honorata", "Tomasz", "Szymon"}
    lastNames := []string{"Kowalski", "Nowak", "Wiśniewski", "Wójcik", "Kowalczyk", "Kamiński", "Lewandowski", "Zieliński", "Szymański", "Woźniak"}
    companies := []string{"TechPol", "InnoSoft", "DataFlow", "CloudBase", "NetSolutions", "DigiMark", "WebCraft", "SoftHouse", "CodeLab", "PixelStudio"}
    statuses := []string{"lead", "prospect", "customer", "inactive"}

    for i := 0; i < 100; i++ {
        fn := firstNames[i%len(firstNames)]
        ln := lastNames[(i/len(firstNames))%len(lastNames)]
        company := companies[i%len(companies)]
        status := statuses[i%len(statuses)]
        email := fmt.Sprintf("%s.%s%d@%s.pl", fn, ln, i, company)
        phone := fmt.Sprintf("+48 %03d %03d %03d", 100+i%900, i%1000, i%1000)

        c := model.Contact{
            ID:      uuid.NewString(),
            Name:    fn + " " + ln,
            Email:   email,
            Phone:   phone,
            Company: company,
            Status:  status,
            Notes:   fmt.Sprintf("Kontakt nr %d", i+1),
        }
        if err := db.Create(&c).Error; err != nil {
            slog.Error("SeedDatabase: failed to create contact", "error", err)
        }
    }
}
