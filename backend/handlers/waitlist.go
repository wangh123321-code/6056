package handlers

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"paper-cutting-workshop/db"
	"paper-cutting-workshop/models"

	"github.com/gin-gonic/gin"
)

const WaitlistConfirmTimeout = 15 * time.Minute

func AddToWaitlist(c *gin.Context) {
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

	if booked < capacity {
		c.JSON(http.StatusBadRequest, gin.H{"error": "课程仍有名额，请直接预约"})
		return
	}

	var existingWaitlistCount int
	err = tx.QueryRow(
		"SELECT COUNT(*) FROM waitlists WHERE course_id = ? AND user_phone = ? AND status IN ('waiting', 'notified')",
		req.CourseID, req.UserPhone,
	).Scan(&existingWaitlistCount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if existingWaitlistCount > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "您已在该课程的候补队列中"})
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

	var existingBookingCount int
	err = tx.QueryRow(
		"SELECT COUNT(*) FROM bookings WHERE course_id = ? AND user_phone = ? AND user_name = ? AND status = 'booked'",
		req.CourseID, req.UserPhone, req.UserName,
	).Scan(&existingBookingCount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if existingBookingCount > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": req.UserName + " 已预约过该课程"})
		return
	}

	var currentMaxPosition int
	err = tx.QueryRow(
		"SELECT COALESCE(MAX(position), 0) FROM waitlists WHERE course_id = ? AND status IN ('waiting', 'notified')",
		req.CourseID,
	).Scan(&currentMaxPosition)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	newPosition := currentMaxPosition + 1

	result, err := tx.Exec(
		"INSERT INTO waitlists (course_id, user_name, user_phone, position, status) VALUES (?, ?, ?, ?, 'waiting')",
		req.CourseID, req.UserName, req.UserPhone, newPosition,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	waitlistID, _ := result.LastInsertId()

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"data": gin.H{
			"id":           waitlistID,
			"course_title": title,
			"date":         date,
			"time_slot":    timeSlot,
			"user_name":    req.UserName,
			"position":     newPosition,
			"type":         "waitlist",
		},
		"message": "已加入候补队列，当前第 " + strconv.Itoa(newPosition) + " 位，有名额释放时将自动通知您",
	})
}

func ConfirmWaitlist(c *gin.Context) {
	var req models.WaitlistConfirmRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误: " + err.Error()})
		return
	}

	tx, err := db.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer tx.Rollback()

	var waitlist models.Waitlist
	var capacity int
	err = tx.QueryRow(`
		SELECT w.id, w.course_id, w.user_name, w.user_phone, w.status, w.expires_at, w.position,
			c.title, c.date, c.time_slot, c.capacity
		FROM waitlists w
		JOIN courses c ON w.course_id = c.id
		WHERE w.id = ?
	`, req.WaitlistID).Scan(
		&waitlist.ID, &waitlist.CourseID, &waitlist.UserName, &waitlist.UserPhone,
		&waitlist.Status, &waitlist.ExpiresAt, &waitlist.Position,
		&waitlist.CourseTitle, &waitlist.CourseDate, &waitlist.CourseSlot, &capacity,
	)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "候补记录不存在"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if waitlist.Status != "notified" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "该候补记录无需确认或已过期"})
		return
	}

	if waitlist.ExpiresAt != nil && waitlist.ExpiresAt.Before(time.Now()) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "确认时间已过期，名额已顺延给下一位"})
		return
	}

	var booked int
	err = tx.QueryRow(
		"SELECT COUNT(*) FROM bookings WHERE course_id = ? AND status = 'booked'",
		waitlist.CourseID,
	).Scan(&booked)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if booked >= capacity {
		c.JSON(http.StatusConflict, gin.H{"error": "名额已被其他用户占用"})
		return
	}

	_, err = tx.Exec(
		"UPDATE waitlists SET status = 'confirmed' WHERE id = ?",
		req.WaitlistID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	result, err := tx.Exec(
		"INSERT INTO bookings (course_id, user_name, user_phone, status) VALUES (?, ?, ?, 'booked')",
		waitlist.CourseID, waitlist.UserName, waitlist.UserPhone,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	bookingID, _ := result.LastInsertId()

	_, err = tx.Exec(`
		UPDATE waitlists
		SET position = position - 1
		WHERE course_id = ? AND status = 'waiting' AND position > (SELECT position FROM waitlists WHERE id = ?)
	`, waitlist.CourseID, req.WaitlistID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"id":           bookingID,
			"course_title": waitlist.CourseTitle,
			"date":         waitlist.CourseDate,
			"time_slot":    waitlist.CourseSlot,
			"user_name":    waitlist.UserName,
		},
		"message": "候补确认成功！",
	})
}

func CancelWaitlist(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	tx, err := db.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer tx.Rollback()

	var status string
	var courseID int64
	var position int
	err = tx.QueryRow("SELECT status, course_id, position FROM waitlists WHERE id = ?", id).Scan(&status, &courseID, &position)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "候补记录不存在"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if status == "cancelled" || status == "confirmed" || status == "expired" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "该候补已处理"})
		return
	}

	_, err = tx.Exec("UPDATE waitlists SET status = 'cancelled' WHERE id = ?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	_, err = tx.Exec(`
		UPDATE waitlists
		SET position = position - 1
		WHERE course_id = ? AND status IN ('waiting', 'notified') AND position > ?
	`, courseID, position)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if status == "notified" {
		err = notifyNextWaitlist(tx, courseID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "已退出候补队列"})
}

func GetCourseWaitlist(c *gin.Context) {
	courseID, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	rows, err := db.DB.Query(`
		SELECT w.id, w.course_id, w.user_name, w.user_phone, w.position, w.status, w.notified_at, w.expires_at, w.created_at
		FROM waitlists w
		WHERE w.course_id = ? AND w.status IN ('waiting', 'notified')
		ORDER BY w.position
	`, courseID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	waitlist := []models.Waitlist{}
	for rows.Next() {
		var w models.Waitlist
		if err := rows.Scan(&w.ID, &w.CourseID, &w.UserName, &w.UserPhone, &w.Position, &w.Status, &w.NotifiedAt, &w.ExpiresAt, &w.CreatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		waitlist = append(waitlist, w)
	}

	c.JSON(http.StatusOK, gin.H{"data": waitlist, "count": len(waitlist)})
}

func notifyNextWaitlist(tx *sql.Tx, courseID int64) error {
	var capacity, booked int
	err := tx.QueryRow(`
		SELECT c.capacity, COALESCE(b.booked_count, 0)
		FROM courses c
		LEFT JOIN (
			SELECT course_id, COUNT(*) as booked_count
			FROM bookings WHERE status = 'booked'
			GROUP BY course_id
		) b ON c.id = b.course_id
		WHERE c.id = ?
	`, courseID).Scan(&capacity, &booked)
	if err != nil {
		return err
	}

	if booked >= capacity {
		return nil
	}

	available := capacity - booked

	for i := 0; i < available; i++ {
		var waitlistID int64
		err = tx.QueryRow(`
			SELECT id FROM waitlists
			WHERE course_id = ? AND status = 'waiting'
			ORDER BY position
			LIMIT 1
		`, courseID).Scan(&waitlistID)
		if err == sql.ErrNoRows {
			return nil
		}
		if err != nil {
			return err
		}

		now := time.Now()
		expiresAt := now.Add(WaitlistConfirmTimeout)

		_, err = tx.Exec(`
			UPDATE waitlists
			SET status = 'notified', notified_at = ?, expires_at = ?
			WHERE id = ?
		`, now, expiresAt, waitlistID)
		if err != nil {
			return err
		}
	}

	return nil
}

func ProcessExpiredWaitlists() {
	tx, err := db.DB.Begin()
	if err != nil {
		return
	}
	defer tx.Rollback()

	rows, err := tx.Query(`
		SELECT w.id, w.course_id
		FROM waitlists w
		WHERE w.status = 'notified' AND w.expires_at < ?
	`, time.Now())
	if err != nil {
		return
	}
	defer rows.Close()

	var expiredIDs []int64
	var courseIDs []int64
	for rows.Next() {
		var id, courseID int64
		if err := rows.Scan(&id, &courseID); err != nil {
			return
		}
		expiredIDs = append(expiredIDs, id)
		courseIDs = append(courseIDs, courseID)
	}

	for i, id := range expiredIDs {
		var position int
		err = tx.QueryRow("SELECT position FROM waitlists WHERE id = ?", id).Scan(&position)
		if err != nil {
			continue
		}

		_, err = tx.Exec("UPDATE waitlists SET status = 'expired' WHERE id = ?", id)
		if err != nil {
			continue
		}

		_, err = tx.Exec(`
			UPDATE waitlists
			SET position = position - 1
			WHERE course_id = ? AND status IN ('waiting', 'notified') AND position > ?
		`, courseIDs[i], position)
		if err != nil {
			continue
		}

		err = notifyNextWaitlist(tx, courseIDs[i])
		if err != nil {
			continue
		}
	}

	tx.Commit()
}
