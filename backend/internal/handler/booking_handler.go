package handler

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"backend/internal/dto"
	"backend/internal/service"
)

type BookingHandler struct {
	svc service.BookingService
}

func NewBookingHandler(svc service.BookingService) *BookingHandler {
	return &BookingHandler{svc: svc}
}

func (h *BookingHandler) GetSettings(c *gin.Context) {
	res, err := h.svc.GetSettings(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *BookingHandler) UpdateSettings(c *gin.Context) {
	userID, _ := c.Get("userID")
	uid, _ := userID.(string)

	var req dto.BookingSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	res, err := h.svc.UpdateSettings(c, uid, req)
	if err != nil {
		if errors.Is(err, service.ErrSlotOutOfWindow) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Nieprawidłowe godziny dostępności"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *BookingHandler) PublicDays(c *gin.Context) {
	from, err1 := time.ParseInLocation("2006-01-02", c.Query("from"), time.Local)
	to, err2 := time.ParseInLocation("2006-01-02", c.Query("to"), time.Local)
	if err1 != nil || err2 != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "from/to wymagane (YYYY-MM-DD)"})
		return
	}
	days, err := h.svc.MonthDays(c, from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"days": days})
}

func (h *BookingHandler) PublicSlots(c *gin.Context) {
	date, err := time.ParseInLocation("2006-01-02", c.Query("date"), time.Local)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "date wymagana (YYYY-MM-DD)"})
		return
	}
	slots, err := h.svc.DaySlots(c, date)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"slots": slots})
}

func (h *BookingHandler) PublicCreate(c *gin.Context) {
	var req dto.BookingCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	res, err := h.svc.CreateBooking(c, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrSlotUnavailable):
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		case errors.Is(err, service.ErrSlotOutOfWindow):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, service.ErrBookingDisabled):
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		}
		return
	}
	c.JSON(http.StatusCreated, res)
}

func (h *BookingHandler) ServeBookingPage(c *gin.Context) {
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(bookingPageHTML))
}

const bookingPageHTML = `<!DOCTYPE html>
<html lang="pl">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Umów spotkanie</title>
<style>
  * { box-sizing: border-box; margin: 0; padding: 0; }
  body { font-family: system-ui, -apple-system, "Segoe UI", sans-serif; background: #0f172a; color: #f1f5f9; min-height: 100vh; display: flex; align-items: flex-start; justify-content: center; padding: 40px 16px; }
  .card { background: #1e293b; border: 1px solid #334155; border-radius: 16px; padding: 28px; width: 100%; max-width: 560px; }
  h1 { font-size: 1.5rem; margin-bottom: 4px; color: #f8fafc; }
  .sub { color: #94a3b8; font-size: 0.9rem; margin-bottom: 22px; }
  .month-nav { display: flex; align-items: center; justify-content: space-between; margin-bottom: 14px; }
  .month-title { font-size: 1.05rem; font-weight: 600; text-transform: capitalize; }
  .nav-btn { background: #0f172a; border: 1px solid #334155; color: #94a3b8; border-radius: 8px; padding: 6px 12px; cursor: pointer; font-size: 1rem; }
  .nav-btn:hover { color: #38bdf8; border-color: #38bdf8; }
  .grid { display: grid; grid-template-columns: repeat(7, 1fr); gap: 6px; }
  .dow { text-align: center; font-size: 0.72rem; color: #64748b; font-weight: 600; padding: 4px 0; }
  .day { aspect-ratio: 1; display: flex; align-items: center; justify-content: center; border-radius: 8px; font-size: 0.9rem; border: 1px solid transparent; }
  .day.empty { visibility: hidden; }
  .day.avail { background: #0f2a4a; color: #38bdf8; border-color: #1e3a5f; cursor: pointer; font-weight: 600; }
  .day.avail:hover { background: #1e3a5f; border-color: #38bdf8; }
  .day.disabled { color: #475569; cursor: not-allowed; }
  .day.selected { background: #38bdf8; color: #0f172a; font-weight: 700; }
  .section-title { margin: 22px 0 12px; font-size: 0.95rem; color: #e2e8f0; }
  .slots { display: grid; grid-template-columns: repeat(auto-fill, minmax(84px, 1fr)); gap: 8px; }
  .slot { padding: 9px 0; text-align: center; border-radius: 8px; border: 1px solid #334155; background: #0f172a; color: #e2e8f0; cursor: pointer; font-size: 0.88rem; }
  .slot:hover { border-color: #38bdf8; color: #38bdf8; }
  .slot.disabled { color: #475569; background: #16202e; cursor: not-allowed; text-decoration: line-through; }
  .slot.selected { background: #38bdf8; color: #0f172a; border-color: #38bdf8; font-weight: 700; }
  .form { margin-top: 22px; display: flex; flex-direction: column; gap: 12px; }
  label { font-size: 0.82rem; color: #94a3b8; display: block; margin-bottom: 5px; }
  input { width: 100%; padding: 10px 12px; background: #0f172a; border: 1px solid #334155; border-radius: 8px; color: #f1f5f9; outline: none; font-size: 0.92rem; }
  input:focus { border-color: #38bdf8; }
  .btn { padding: 12px; background: #38bdf8; border: none; border-radius: 8px; color: #0f172a; font-weight: 700; cursor: pointer; font-size: 0.95rem; }
  .btn:disabled { opacity: 0.5; cursor: not-allowed; }
  .muted { color: #64748b; font-size: 0.85rem; padding: 10px 0; }
  .err { color: #f87171; font-size: 0.86rem; }
  .ok { text-align: center; padding: 20px 0; }
  .ok .check { font-size: 3rem; }
  .ok h2 { color: #34d399; margin: 10px 0 6px; }
  .ok p { color: #94a3b8; }
  .hidden { display: none; }
  .pick-note { color: #64748b; font-size: 0.85rem; }
</style>
</head>
<body>
<div class="card">
  <div id="booking">
    <h1>Umów spotkanie</h1>
    <p class="sub">Wybierz dogodny dzień i godzinę.</p>

    <div class="month-nav">
      <button class="nav-btn" id="prev">&#8249;</button>
      <div class="month-title" id="monthTitle"></div>
      <button class="nav-btn" id="next">&#8250;</button>
    </div>
    <div class="grid" id="dow"></div>
    <div class="grid" id="days"></div>

    <div id="slotsSection" class="hidden">
      <div class="section-title" id="slotsTitle"></div>
      <div class="slots" id="slots"></div>
      <p class="pick-note" id="noSlots"></p>
    </div>

    <div id="formSection" class="form hidden">
      <div class="section-title">Twoje dane</div>
      <div>
        <label>Imię i nazwisko *</label>
        <input id="fName" placeholder="Jan Kowalski" />
      </div>
      <div>
        <label>E-mail</label>
        <input id="fEmail" type="email" placeholder="jan@example.com" />
      </div>
      <div>
        <label>Telefon</label>
        <input id="fPhone" placeholder="+48 600 000 000" />
      </div>
      <p class="err hidden" id="formErr"></p>
      <button class="btn" id="submitBtn">Umów się</button>
    </div>
  </div>

  <div id="success" class="ok hidden">
    <div class="check">&#9989;</div>
    <h2>Termin zarezerwowany!</h2>
    <p id="successMsg"></p>
  </div>
</div>

<script>
(function () {
  "use strict";
  var DOW = ["Pn", "Wt", "Sr", "Cz", "Pt", "Sb", "Nd"];
  var MONTHS = ["Styczen", "Luty", "Marzec", "Kwiecien", "Maj", "Czerwiec", "Lipiec", "Sierpien", "Wrzesien", "Pazdziernik", "Listopad", "Grudzien"];

  var view = new Date();
  view.setDate(1);
  var selectedDate = null;
  var selectedSlot = null;

  var dowEl = document.getElementById("dow");
  var daysEl = document.getElementById("days");
  var monthTitle = document.getElementById("monthTitle");
  var slotsSection = document.getElementById("slotsSection");
  var slotsEl = document.getElementById("slots");
  var slotsTitle = document.getElementById("slotsTitle");
  var noSlots = document.getElementById("noSlots");
  var formSection = document.getElementById("formSection");
  var formErr = document.getElementById("formErr");
  var submitBtn = document.getElementById("submitBtn");

  for (var i = 0; i < DOW.length; i++) {
    var d = document.createElement("div");
    d.className = "dow";
    d.textContent = DOW[i];
    dowEl.appendChild(d);
  }

  function pad(n) { return n < 10 ? "0" + n : "" + n; }
  function ymd(date) { return date.getFullYear() + "-" + pad(date.getMonth() + 1) + "-" + pad(date.getDate()); }

  function resetSelection() {
    selectedDate = null;
    selectedSlot = null;
    slotsSection.classList.add("hidden");
    formSection.classList.add("hidden");
    slotsEl.innerHTML = "";
    noSlots.textContent = "";
  }

  document.getElementById("prev").addEventListener("click", function () {
    view.setMonth(view.getMonth() - 1);
    resetSelection();
    loadMonth();
  });
  document.getElementById("next").addEventListener("click", function () {
    view.setMonth(view.getMonth() + 1);
    resetSelection();
    loadMonth();
  });

  function loadMonth() {
    var year = view.getFullYear();
    var month = view.getMonth();
    monthTitle.textContent = MONTHS[month] + " " + year;
    var first = new Date(year, month, 1);
    var last = new Date(year, month + 1, 0);

    fetch("/public/v1/booking/days?from=" + ymd(first) + "&to=" + ymd(last))
      .then(function (r) { return r.json(); })
      .then(function (data) { renderDays(first, last, (data && data.days) || []); })
      .catch(function () { renderDays(first, last, []); });
  }

  function renderDays(first, last, days) {
    var avail = {};
    for (var i = 0; i < days.length; i++) { avail[days[i].date] = days[i].available; }

    daysEl.innerHTML = "";
    var startDow = first.getDay() === 0 ? 7 : first.getDay(); // Monday-first
    for (var e = 1; e < startDow; e++) {
      var blank = document.createElement("div");
      blank.className = "day empty";
      daysEl.appendChild(blank);
    }
    for (var dnum = 1; dnum <= last.getDate(); dnum++) {
      var date = new Date(first.getFullYear(), first.getMonth(), dnum);
      var key = ymd(date);
      var cell = document.createElement("div");
      cell.textContent = dnum;
      if (avail[key]) {
        cell.className = "day avail";
        cell.dataset.date = key;
        cell.addEventListener("click", onDayClick);
      } else {
        cell.className = "day disabled";
      }
      daysEl.appendChild(cell);
    }
  }

  function onDayClick(ev) {
    var key = ev.currentTarget.dataset.date;
    selectedDate = key;
    selectedSlot = null;
    formSection.classList.add("hidden");
    var cells = daysEl.querySelectorAll(".day");
    for (var i = 0; i < cells.length; i++) { cells[i].classList.remove("selected"); }
    ev.currentTarget.classList.add("selected");
    loadSlots(key);
  }

  function loadSlots(key) {
    slotsSection.classList.remove("hidden");
    slotsTitle.textContent = "Godziny - " + key;
    slotsEl.innerHTML = "";
    noSlots.textContent = "Ladowanie...";
    fetch("/public/v1/booking/slots?date=" + key)
      .then(function (r) { return r.json(); })
      .then(function (data) { renderSlots((data && data.slots) || []); })
      .catch(function () { renderSlots([]); });
  }

  function renderSlots(slots) {
    slotsEl.innerHTML = "";
    var anyAvail = false;
    for (var i = 0; i < slots.length; i++) {
      var s = slots[i];
      var btn = document.createElement("div");
      btn.textContent = s.start;
      if (s.available) {
        anyAvail = true;
        btn.className = "slot";
        btn.dataset.start = s.start;
        btn.addEventListener("click", onSlotClick);
      } else {
        btn.className = "slot disabled";
      }
      slotsEl.appendChild(btn);
    }
    noSlots.textContent = slots.length === 0 ? "Brak godzin w tym dniu." : (anyAvail ? "" : "Wszystkie godziny zajete.");
  }

  function onSlotClick(ev) {
    selectedSlot = ev.currentTarget.dataset.start;
    var btns = slotsEl.querySelectorAll(".slot");
    for (var i = 0; i < btns.length; i++) { btns[i].classList.remove("selected"); }
    ev.currentTarget.classList.add("selected");
    formSection.classList.remove("hidden");
    formErr.classList.add("hidden");
  }

  submitBtn.addEventListener("click", function () {
    var name = document.getElementById("fName").value.trim();
    if (!name) { showErr("Podaj imie i nazwisko."); return; }
    if (!selectedDate || !selectedSlot) { showErr("Wybierz dzien i godzine."); return; }

    submitBtn.disabled = true;
    submitBtn.textContent = "Rezerwowanie...";

    fetch("/public/v1/booking", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        date: selectedDate,
        start: selectedSlot,
        name: name,
        email: document.getElementById("fEmail").value.trim(),
        phone: document.getElementById("fPhone").value.trim()
      })
    }).then(function (r) {
      return r.json().then(function (body) { return { ok: r.ok, body: body }; });
    }).then(function (res) {
      if (!res.ok) {
        showErr((res.body && res.body.error) || "Nie udalo sie zarezerwowac terminu.");
        submitBtn.disabled = false;
        submitBtn.textContent = "Umow sie";
        if (selectedDate) { loadSlots(selectedDate); } // refresh in case slot got taken
        return;
      }
      document.getElementById("booking").classList.add("hidden");
      var ok = document.getElementById("success");
      ok.classList.remove("hidden");
      document.getElementById("successMsg").textContent = "Spotkanie " + selectedDate + " o " + selectedSlot + " zostalo umowione.";
    }).catch(function () {
      showErr("Blad polaczenia. Sprobuj ponownie.");
      submitBtn.disabled = false;
      submitBtn.textContent = "Umow sie";
    });
  });

  function showErr(msg) {
    formErr.textContent = msg;
    formErr.classList.remove("hidden");
  }

  loadMonth();
})();
</script>
</body>
</html>`
