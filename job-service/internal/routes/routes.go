package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/kidus-yoseph1/job-board-platform/job-service/internal/handler"
	"github.com/kidus-yoseph1/job-board-platform/job-service/internal/middleware"
)

func Setup(
	r *gin.Engine,
	applicationHandler *handler.ApplicationHandler,
	jobHandler *handler.JobHandler,
	authHandler *handler.AuthHandler,
	jwtSecret string,
) {
	//public
	r.GET("/jobs", jobHandler.ListJobsHandler)
	r.GET("/job/:id", jobHandler.GetJobByIDHandler)
	r.GET("/jobs/company/:id", jobHandler.ListJobsByCompanyHandler)
	r.POST("/auth/register", authHandler.RegisterHandler)
	r.POST("/auth/login", authHandler.LoginHandler)

	// protected
	protected := r.Group("/")
	protected.Use(middleware.AuthMiddleware(jwtSecret))
	protected.Use(middleware.RequireRole("candidate"))
	{
		// applications
		protected.POST("/apply/:id", applicationHandler.CreateApplicationHandler)
		protected.GET("/application/:id", applicationHandler.GetApplicationByIDHandler)
		protected.GET("/application", applicationHandler.ListApplicationsByUserHandler)
	}

	//company
	company := r.Group("/")
	company.Use(middleware.AuthMiddleware(jwtSecret))
	company.Use(middleware.RequireRole("company"))

	{
		// jobs
		company.POST("/create_job", jobHandler.CreateJobHandler)
		company.PUT("/update_job/:id", jobHandler.UpdateJobStatusHandler)
		company.DELETE("/delete_job/:id", jobHandler.DeleteJobHandler)

		// applications
		company.GET("/application/job/:job_id", applicationHandler.ListApplicationsByJobHandler)
		company.PUT("/application/update/:id", applicationHandler.UpdateApplicationStatusHandler)
	}
}
