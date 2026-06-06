package handlers

import (
	"database/sql"
	"net/http"
	"strconv"

	"paper-cutting-workshop/db"
	"paper-cutting-workshop/models"

	"github.com/gin-gonic/gin"
)

func GetCourses(c *gin.Context) {
	date := c.Query("date")
	month := c.Query("month")

	query := `
		SELECT c.id, c.title, c.date, c.time_slot, c.capacity, c.created_at, c.updated_at,
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

	if date != "" {
		query += " WHERE c.date = ? ORDER BY c.date, c.time_slot"
		rows, err = db.DB.Query(query, date)
	} else if month != "" {
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

	courses := []models.Course{}
	for rows.Next() {
		var co models.Course
		if err := rows.Scan(&co.ID, &co.Title, &co.Date, &co.TimeSlot, &co.Capacity, &co.CreatedAt, &co.UpdatedAt, &co.Booked); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		courses = append(courses, co)
	}

	c.JSON(http.StatusOK, gin.H{"data": courses})
}

func GetCourse(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var co models.Course
	err := db.DB.QueryRow(`
		SELECT c.id, c.title, c.date, c.time_slot, c.capacity, c.created_at, c.updated_at,
			COALESCE(b.booked_count, 0) as booked
		FROM courses c
		LEFT JOIN (
			SELECT course_id, COUNT(*) as booked_count
			FROM bookings WHERE status = 'booked'
			GROUP BY course_id
		) b ON c.id = b.course_id
		WHERE c.id = ?
	`, id).Scan(&co.ID, &co.Title, &co.Date, &co.TimeSlot, &co.Capacity, &co.CreatedAt, &co.UpdatedAt, &co.Booked)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "课程不存在"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": co})
}

func CreateCourse(c *gin.Context) {
	var req models.CourseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误: " + err.Error()})
		return
	}

	result, err := db.DB.Exec(
		"INSERT INTO courses (title, date, time_slot, capacity) VALUES (?, ?, ?, ?)",
		req.Title, req.Date, req.TimeSlot, req.Capacity,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	id, _ := result.LastInsertId()
	c.JSON(http.StatusCreated, gin.H{"data": gin.H{"id": id}, "message": "课程创建成功"})
}

func UpdateCourse(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var req models.CourseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误: " + err.Error()})
		return
	}

	var booked int
	err := db.DB.QueryRow("SELECT COUNT(*) FROM bookings WHERE course_id = ? AND status = 'booked'", id).Scan(&booked)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if req.Capacity < booked {
		c.JSON(http.StatusBadRequest, gin.H{"error": "名额不能少于已预约人数(" + strconv.Itoa(booked) + "人)"})
		return
	}

	result, err := db.DB.Exec(
		"UPDATE courses SET title=?, date=?, time_slot=?, capacity=?, updated_at=CURRENT_TIMESTAMP WHERE id=?",
		req.Title, req.Date, req.TimeSlot, req.Capacity, id,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "课程不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "课程更新成功"})
}

func DeleteCourse(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var booked int
	err := db.DB.QueryRow("SELECT COUNT(*) FROM bookings WHERE course_id = ? AND status = 'booked'", id).Scan(&booked)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if booked > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "该课程还有" + strconv.Itoa(booked) + "人预约，无法删除"})
		return
	}

	result, err := db.DB.Exec("DELETE FROM courses WHERE id=?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "课程不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "课程删除成功"})
}
