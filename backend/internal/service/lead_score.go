package service

import (
    "time"

    "backend/internal/model"
)

func CalcLeadScore(c *model.Contact, historyCount int, lastActivityDays int) int {
    score := 0

    switch c.Status {
    case "customer":
        score += 40
    case "prospect":
        score += 25
    case "lead":
        score += 10
    case "inactive":
        score -= 10
    }

    if c.Email != "" {
        score += 10
    }
    if c.Phone != "" {
        score += 10
    }
    if c.Company != "" {
        score += 10
    }
    if c.Notes != "" {
        score += 5
    }

    if historyCount >= 10 {
        score += 15
    } else if historyCount >= 5 {
        score += 10
    } else if historyCount >= 2 {
        score += 5
    }

    if lastActivityDays >= 0 {
        if lastActivityDays <= 7 {
            score += 10
        } else if lastActivityDays <= 30 {
            score += 5
        } else if lastActivityDays > 90 {
            score -= 10
        }
    }

    if score < 0 {
        score = 0
    }
    if score > 100 {
        score = 100
    }
    return score
}

func DaysSince(t time.Time) int {
    return int(time.Since(t).Hours() / 24)
}

func calcLeadScore(c *model.Contact, historyCount int, lastActivityDays int) int {
    return CalcLeadScore(c, historyCount, lastActivityDays)
}

func daysSince(t time.Time) int {
    return DaysSince(t)
}
