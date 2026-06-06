package handlers

import (
	"database/sql"
	"math"
	"net/http"
	"regexp"
	"strconv"

	"paper-cutting-workshop/db"
	"paper-cutting-workshop/models"

	"github.com/gin-gonic/gin"
)

var phoneRegex = regexp.MustCompile(`^1[3-9]\d{9}$`)

func CreateBooking(c *gin.Context) {
	var req models.BookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误: " + err.Error()})
		return
	}

	if !phoneRegex.MatchString(req.UserPhone) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请输入正确的11位手机号"})
		return
	}

	tx, err := db.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer tx.Rollback()

	var capacity, booked int
	var title, date, timeSlot string
	err = tx.QueryRow(
		"SELECT capacity, title, date, time_slot FROM courses WHERE id = ?",
		req.CourseID,
	).Scan(&capacity, &title, &date, &timeSlot)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "课程不存在"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	err = tx.QueryRow(
		"SELECT COUNT(*) FROM bookings WHERE course_id = ? AND status = 'booked'",
		req.CourseID,
	).Scan(&booked)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if booked >= capacity {
		c.JSON(http.StatusConflict, gin.H{"error": "名额已满，无法预约"})
		return
	}

	var phoneBookingCount int
	err = tx.QueryRow(
		"SELECT COUNT(*) FROM bookings WHERE course_id = ? AND user_phone = ? AND status = 'booked'",
		req.CourseID, req.UserPhone,
	).Scan(&phoneBookingCount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if phoneBookingCount >= 3 {
		c.JSON(http.StatusConflict, gin.H{"error": "同一手机号最多预约3个名额，如需更多请联系管理员"})
		return
	}

	var existingCount int
	err = tx.QueryRow(
		"SELECT COUNT(*) FROM bookings WHERE course_id = ? AND user_phone = ? AND user_name = ? AND status = 'booked'",
		req.CourseID, req.UserPhone, req.UserName,
	).Scan(&existingCount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if existingCount > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": req.UserName + " 已预约过该课程"})
		return
	}

	result, err := tx.Exec(
		"INSERT INTO bookings (course_id, user_name, user_phone, status) VALUES (?, ?, ?, 'booked')",
		req.CourseID, req.UserName, req.UserPhone,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	bookingID, _ := result.LastInsertId()

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"data": gin.H{
			"id":           bookingID,
			"course_title": title,
			"date":         date,
			"time_slot":    timeSlot,
			"user_name":    req.UserName,
			"remaining":    capacity - booked - 1,
		},
		"message": "预约成功！",
	})
}

func CancelBooking(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	tx, err := db.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer tx.Rollback()

	var status string
	err = tx.QueryRow("SELECT status FROM bookings WHERE id = ?", id).Scan(&status)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "预约记录不存在"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if status == "cancelled" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "该预约已取消"})
		return
	}

	_, err = tx.Exec(
		"UPDATE bookings SET status = 'cancelled', cancelled_at = CURRENT_TIMESTAMP WHERE id = ?",
		id,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "预约已取消，名额已释放"})
}

func GetMyBookings(c *gin.Context) {
	phone := c.Query("phone")
	if phone == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请提供手机号"})
		return
	}

	rows, err := db.DB.Query(`
		SELECT b.id, b.course_id, b.user_name, b.user_phone, b.status, b.created_at, b.cancelled_at,
			c.title as course_title, c.date as course_date, c.time_slot as course_slot
		FROM bookings b
		JOIN courses c ON b.course_id = c.id
		WHERE b.user_phone = ?
		ORDER BY b.created_at DESC
	`, phone)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	bookings := []models.Booking{}
	for rows.Next() {
		var b models.Booking
		if err := rows.Scan(&b.ID, &b.CourseID, &b.UserName, &b.UserPhone, &b.Status, &b.CreatedAt, &b.CancelledAt, &b.CourseTitle, &b.CourseDate, &b.CourseSlot); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		bookings = append(bookings, b)
	}

	c.JSON(http.StatusOK, gin.H{"data": bookings})
}

func GetCourseBookings(c *gin.Context) {
	courseID, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	rows, err := db.DB.Query(`
		SELECT b.id, b.course_id, b.user_name, b.user_phone, b.status, b.created_at, b.cancelled_at
		FROM bookings b
		WHERE b.course_id = ?
		ORDER BY b.created_at
	`, courseID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	bookings := []models.Booking{}
	for rows.Next() {
		var b models.Booking
		if err := rows.Scan(&b.ID, &b.CourseID, &b.UserName, &b.UserPhone, &b.Status, &b.CreatedAt, &b.CancelledAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		bookings = append(bookings, b)
	}

	c.JSON(http.StatusOK, gin.H{"data": bookings})
}

func CheckAvailability(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var capacity, booked int
	err := db.DB.QueryRow(`
		SELECT c.capacity, COALESCE(b.booked_count, 0)
		FROM courses c
		LEFT JOIN (
			SELECT course_id, COUNT(*) as booked_count
			FROM bookings WHERE status = 'booked'
			GROUP BY course_id
		) b ON c.id = b.course_id
		WHERE c.id = ?
	`, id).Scan(&capacity, &booked)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "课程不存在"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	available := capacity - booked
	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"course_id":  id,
			"capacity":   capacity,
			"booked":     booked,
			"available":  available,
			"is_full":    available <= 0,
		},
	})
}

func GetCalendarEvents(c *gin.Context) {
	month := c.Query("month")

	query := `
		SELECT c.id, c.title, c.date, c.time_slot, c.capacity,
			COALESCE(b.booked_count, 0) as booked
		FROM courses c
		LEFT JOIN (
			SELECT course_id, COUNT(*) as booked_count
			FROM bookings WHERE status = 'booked'
			GROUP BY course_id
		) b ON c.id = b.course_id
	`
	var rows *sql.Rows
	var err error

	if month != "" {
		query += " WHERE strftime('%Y-%m', c.date) = ? ORDER BY c.date, c.time_slot"
		rows, err = db.DB.Query(query, month)
	} else {
		query += " ORDER BY c.date, c.time_slot"
		rows, err = db.DB.Query(query)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	type CalendarEvent struct {
		ID       int64  `json:"id"`
		Title    string `json:"title"`
		Date     string `json:"date"`
		TimeSlot string `json:"time_slot"`
		Capacity int    `json:"capacity"`
		Booked   int    `json:"booked"`
		IsFull   bool   `json:"is_full"`
	}

	events := []CalendarEvent{}
	for rows.Next() {
		var e CalendarEvent
		if err := rows.Scan(&e.ID, &e.Title, &e.Date, &e.TimeSlot, &e.Capacity, &e.Booked); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		e.IsFull = e.Booked >= e.Capacity
		events = append(events, e)
	}

	c.JSON(http.StatusOK, gin.H{"data": events})
}

func GetStats(c *gin.Context) {
	query := `
		SELECT c.id, c.title, c.date, c.time_slot, c.capacity,
			COALESCE(b.booked_count, 0),
			COALESCE(a.attended_count, 0)
		FROM courses c
		LEFT JOIN (
			SELECT course_id, COUNT(*) as booked_count
			FROM bookings WHERE status = 'booked'
			GROUP BY course_id
		) b ON c.id = b.course_id
		LEFT JOIN (
			SELECT course_id, COUNT(*) as attended_count
			FROM bookings WHERE status = 'attended'
			GROUP BY course_id
		) a ON c.id = a.course_id
		ORDER BY c.date DESC, c.time_slot
	`

	rows, err := db.DB.Query(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	stats := []models.StatsResponse{}
	for rows.Next() {
		var s models.StatsResponse
		if err := rows.Scan(&s.CourseID, &s.CourseTitle, &s.CourseDate, &s.CourseSlot, &s.Capacity, &s.Booked, &s.Attended); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if s.Booked > 0 {
			s.AttendRate = math.Round(float64(s.Attended)/float64(s.Booked)*1000) / 10
		} else {
			s.AttendRate = 0.0
		}
		stats = append(stats, s)
	}

	c.JSON(http.StatusOK, gin.H{"data": stats})
}

func MarkAttendance(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	result, err := db.DB.Exec("UPDATE bookings SET status = 'attended' WHERE id = ? AND status = 'booked'", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "预约不存在或已处理"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "已标记到课"})
}
