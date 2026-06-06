package main

import (
	"log"

	"paper-cutting-workshop/db"
	"paper-cutting-workshop/handlers"
	"paper-cutting-workshop/middleware"

	"github.com/gin-gonic/gin"
)

func main() {
	db.Init("/app/data/workshop.db")

	r := gin.Default()
	r.Use(middleware.CORS())

	api := r.Group("/api")
	{
		courses := api.Group("/courses")
		{
			courses.GET("", handlers.GetCourses)
			courses.GET("/calendar", handlers.GetCalendarEvents)
			courses.GET("/:id", handlers.GetCourse)
			courses.POST("", handlers.CreateCourse)
			courses.PUT("/:id", handlers.UpdateCourse)
			courses.DELETE("/:id", handlers.DeleteCourse)
		}

		bookings := api.Group("/bookings")
		{
			bookings.POST("", handlers.CreateBooking)
			bookings.GET("/my", handlers.GetMyBookings)
			bookings.DELETE("/:id", handlers.CancelBooking)
		}

		courseBookings := api.Group("/courses")
		{
			courseBookings.GET("/:id/bookings", handlers.GetCourseBookings)
			courseBookings.GET("/:id/availability", handlers.CheckAvailability)
		}

		stats := api.Group("/stats")
		{
			stats.GET("", handlers.GetStats)
		}

		attendance := api.Group("/attendance")
		{
			attendance.PUT("/:id", handlers.MarkAttendance)
		}
	}

	log.Println("Server starting on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
