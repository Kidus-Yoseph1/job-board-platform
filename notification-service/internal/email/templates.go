package email

import (
	"bytes"
	"fmt"
	"html/template"
)

// Data structures for email templates
type WelcomeData struct {
	FullName string
	Role     string
}

type JobAppliedData struct {
	EmployerName  string
	CandidateName string
	JobTitle      string
	CoverLetter   string
}

type ApplicationStatusData struct {
	CandidateName string
	JobTitle      string
	CompanyName   string
	Status        string
}

const (
	WelcomeTemplate           = "welcome"
	JobAppliedTemplate        = "job_applied"
	ApplicationStatusTemplate = "application_status"
)

var templates = make(map[string]*template.Template)

func init() {
	// 1. Welcome template
	templates[WelcomeTemplate] = template.Must(template.New(WelcomeTemplate).Parse(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; border: 1px solid #dddddd; border-radius: 5px; }
        .header { background-color: #4A90E2; padding: 15px; text-align: center; color: white; border-radius: 5px 5px 0 0; }
        .content { padding: 20px; }
        .footer { text-align: center; font-size: 12px; color: #777777; margin-top: 20px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h2>Welcome to Job Board Platform!</h2>
        </div>
        <div class="content">
            <p>Hi {{.FullName}},</p>
            <p>Thank you for registering on our platform as a <strong>{{.Role}}</strong>. We are thrilled to have you join our community!</p>
            <p>Start exploring available jobs or manage your listings now.</p>
        </div>
        <div class="footer">
            <p>&copy; 2026 Job Board Platform. All rights reserved.</p>
        </div>
    </div>
</body>
</html>
`))

	// 2. Job Applied template (to Employer)
	templates[JobAppliedTemplate] = template.Must(template.New(JobAppliedTemplate).Parse(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; border: 1px solid #dddddd; border-radius: 5px; }
        .header { background-color: #4A90E2; padding: 15px; text-align: center; color: white; border-radius: 5px 5px 0 0; }
        .content { padding: 20px; }
        .cover-letter { background-color: #f9f9f9; border-left: 4px solid #4A90E2; padding: 10px; margin: 15px 0; font-style: italic; }
        .footer { text-align: center; font-size: 12px; color: #777777; margin-top: 20px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h2>New Job Application Received</h2>
        </div>
        <div class="content">
            <p>Hi {{.EmployerName}},</p>
            <p>Great news! <strong>{{.CandidateName}}</strong> has applied for your job post: <strong>{{.JobTitle}}</strong>.</p>
            {{if .CoverLetter}}
            <p><strong>Cover Letter:</strong></p>
            <div class="cover-letter">
                {{.CoverLetter}}
            </div>
            {{end}}
            <p>Please log in to your dashboard to review this application.</p>
        </div>
        <div class="footer">
            <p>&copy; 2026 Job Board Platform. All rights reserved.</p>
        </div>
    </div>
</body>
</html>
`))

	// 3. Application Status template (to Candidate)
	templates[ApplicationStatusTemplate] = template.Must(template.New(ApplicationStatusTemplate).Parse(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; border: 1px solid #dddddd; border-radius: 5px; }
        .header { background-color: #4A90E2; padding: 15px; text-align: center; color: white; border-radius: 5px 5px 0 0; }
        .content { padding: 20px; }
        .status-badge { display: inline-block; padding: 5px 10px; background-color: #E2F0D9; color: #385723; font-weight: bold; border-radius: 3px; }
        .footer { text-align: center; font-size: 12px; color: #777777; margin-top: 20px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h2>Application Update</h2>
        </div>
        <div class="content">
            <p>Hi {{.CandidateName}},</p>
            <p>Your application status for the job <strong>{{.JobTitle}}</strong> at <strong>{{.CompanyName}}</strong> has been updated:</p>
            <p>New Status: <span class="status-badge">{{.Status}}</span></p>
            <p>Please log in to your dashboard to view more details.</p>
        </div>
        <div class="footer">
            <p>&copy; 2026 Job Board Platform. All rights reserved.</p>
        </div>
    </div>
</body>
</html>
`))
}

// RenderTemplate renders the template with the provided data and returns the HTML string
func RenderTemplate(templateName string, data interface{}) (string, error) {
	tmpl, exists := templates[templateName]
	if !exists {
		return "", fmt.Errorf("template not found: %s", templateName)
	}

	var buf bytes.Buffer
	err := tmpl.Execute(&buf, data)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}
