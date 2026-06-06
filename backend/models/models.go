package models

import "time"

type Course struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	Date      string    `json:"date"`
	TimeSlot  string    `json:"time_slot"`
	Capacity  int       `json:"capacity"`
	Booked    int       `json:"booked"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Booking struct {
	ID          int64      `json:"id"`
	CourseID    int64      `json:"course_id"`
	UserName    string     `json:"user_name"`
	UserPhone   string     `json:"user_phone"`
	Status      string     `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
	CancelledAt *time.Time `json:"cancelled_at,omitempty"`
	CourseTitle string     `json:"course_title,omitempty"`
	CourseDate  string     `json:"course_date,omitempty"`
	CourseSlot  string     `json:"course_slot,omitempty"`
}

type BookingRequest struct {
	CourseID  int64  `json:"course_id" binding:"required"`
	UserName  string `json:"user_name" binding:"required"`
	UserPhone string `json:"user_phone" binding:"required"`
}

type CourseRequest struct {
	Title    string `json:"title" binding:"required"`
	Date     string `json:"date" binding:"required"`
	TimeSlot string `json:"time_slot" binding:"required"`
	Capacity int    `json:"capacity" binding:"required,min=1"`
}

type StatsResponse struct {
	CourseID    int64   `json:"course_id"`
	CourseTitle string  `json:"course_title"`
	CourseDate  string  `json:"course_date"`
	CourseSlot  string  `json:"course_slot"`
	Capacity    int     `json:"capacity"`
	Booked      int     `json:"booked"`
	Attended    int     `json:"attended"`
	AttendRate  float64 `json:"attend_rate"`
}
