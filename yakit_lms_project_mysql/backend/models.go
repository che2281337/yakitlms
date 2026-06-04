package main

import (
	"time"
)

type User struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	FullName     string    `gorm:"size:160;not null" json:"fullName"`
	Email        string    `gorm:"size:160;uniqueIndex;not null" json:"email"`
	PasswordHash string    `gorm:"not null" json:"-"`
	Role         string    `gorm:"size:24;index;not null;default:student" json:"role"`
	GroupName    string    `gorm:"size:80" json:"groupName"`
	Specialty    string    `gorm:"size:220" json:"specialty"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

type Course struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Title       string         `gorm:"size:180;not null" json:"title"`
	Category    string         `gorm:"size:80" json:"category"`
	Description string         `gorm:"type:text" json:"description"`
	ImageURL    string         `gorm:"size:500" json:"imageUrl"`
	TeacherID   uint           `gorm:"index" json:"teacherId"`
	Teacher     User           `gorm:"foreignKey:TeacherID" json:"teacher"`
	Status      string         `gorm:"size:40;default:active" json:"status"`
	Modules     []CourseModule `gorm:"foreignKey:CourseID" json:"modules,omitempty"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
}

type Enrollment struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"uniqueIndex:idx_enrollment" json:"userId"`
	User      User      `json:"user,omitempty"`
	CourseID  uint      `gorm:"uniqueIndex:idx_enrollment" json:"courseId"`
	Progress  int       `gorm:"default:0" json:"progress"`
	CreatedAt time.Time `json:"createdAt"`
}

type CourseModule struct {
	ID          uint         `gorm:"primaryKey" json:"id"`
	CourseID    uint         `gorm:"index;not null" json:"courseId"`
	Title       string       `gorm:"size:180;not null" json:"title"`
	Description string       `gorm:"type:text" json:"description"`
	SortOrder   int          `gorm:"default:1" json:"sortOrder"`
	Materials   []Material   `gorm:"foreignKey:ModuleID" json:"materials,omitempty"`
	Assignments []Assignment `gorm:"foreignKey:ModuleID" json:"assignments,omitempty"`
	Tests       []Test       `gorm:"foreignKey:ModuleID" json:"tests,omitempty"`
	CreatedAt   time.Time    `json:"createdAt"`
	UpdatedAt   time.Time    `json:"updatedAt"`
}

type Material struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CourseID  uint      `gorm:"index;not null" json:"courseId"`
	ModuleID  uint      `gorm:"index;not null" json:"moduleId"`
	Title     string    `gorm:"size:180;not null" json:"title"`
	Kind      string    `gorm:"size:40;default:link" json:"kind"`
	URL       string    `gorm:"size:500" json:"url"`
	Content   string    `gorm:"type:text" json:"content"`
	SortOrder int       `gorm:"default:1" json:"sortOrder"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type Assignment struct {
	ID             uint         `gorm:"primaryKey" json:"id"`
	CourseID       uint         `gorm:"index;not null" json:"courseId"`
	ModuleID       uint         `gorm:"index;not null" json:"moduleId"`
	Course         Course       `gorm:"foreignKey:CourseID" json:"course,omitempty"`
	Module         CourseModule `gorm:"foreignKey:ModuleID" json:"module,omitempty"`
	Title          string       `gorm:"size:180;not null" json:"title"`
	Description    string       `gorm:"type:text" json:"description"`
	InstructionURL string       `gorm:"size:500" json:"instructionUrl"`
	Kind           string       `gorm:"size:40;default:assignment" json:"kind"`
	Deadline       time.Time    `json:"deadline"`
	MaxScore       int          `gorm:"default:5" json:"maxScore"`
	CreatedAt      time.Time    `json:"createdAt"`
	UpdatedAt      time.Time    `json:"updatedAt"`
}

type Submission struct {
	ID           uint             `gorm:"primaryKey" json:"id"`
	AssignmentID uint             `gorm:"index;not null" json:"assignmentId"`
	Assignment   Assignment       `gorm:"foreignKey:AssignmentID" json:"assignment,omitempty"`
	StudentID    uint             `gorm:"index;not null" json:"studentId"`
	Student      User             `gorm:"foreignKey:StudentID" json:"student,omitempty"`
	Answer       string           `gorm:"type:text" json:"answer"`
	FileURL      string           `gorm:"size:500" json:"fileUrl"`
	FileName     string           `gorm:"size:260" json:"fileName"`
	Files        []SubmissionFile `gorm:"foreignKey:SubmissionID" json:"files,omitempty"`
	Status       string           `gorm:"size:40;default:pending" json:"status"`
	Grade        *int             `json:"grade"`
	Comment      string           `gorm:"type:text" json:"comment"`
	CreatedAt    time.Time        `json:"createdAt"`
	UpdatedAt    time.Time        `json:"updatedAt"`
}

type SubmissionFile struct {
	ID           uint       `gorm:"primaryKey" json:"id"`
	SubmissionID uint       `gorm:"index;not null" json:"submissionId"`
	Submission   Submission `gorm:"foreignKey:SubmissionID" json:"-"`
	FileURL      string     `gorm:"size:500;not null" json:"fileUrl"`
	FileName     string     `gorm:"size:260" json:"fileName"`
	CreatedAt    time.Time  `json:"createdAt"`
}

type Test struct {
	ID              uint           `gorm:"primaryKey" json:"id"`
	CourseID        uint           `gorm:"index;not null" json:"courseId"`
	ModuleID        uint           `gorm:"index;not null" json:"moduleId"`
	Course          Course         `gorm:"foreignKey:CourseID" json:"course,omitempty"`
	Module          CourseModule   `gorm:"foreignKey:ModuleID" json:"module,omitempty"`
	Title           string         `gorm:"size:180;not null" json:"title"`
	Description     string         `gorm:"type:text" json:"description"`
	Deadline        time.Time      `json:"deadline"`
	MaxScore        int            `gorm:"default:5" json:"maxScore"`
	AttemptLimit    int            `gorm:"default:1" json:"attemptLimit"`
	DurationMinutes int            `gorm:"default:0" json:"durationMinutes"`
	ShowResult      bool           `gorm:"default:true" json:"showResult"`
	Questions       []TestQuestion `gorm:"foreignKey:TestID" json:"questions,omitempty"`
	CreatedAt       time.Time      `json:"createdAt"`
	UpdatedAt       time.Time      `json:"updatedAt"`
}

type TestQuestion struct {
	ID                 uint      `gorm:"primaryKey" json:"id"`
	TestID             uint      `gorm:"index;not null" json:"testId"`
	Question           string    `gorm:"type:text;not null" json:"question"`
	Type               string    `gorm:"size:40;default:radio" json:"type"`
	ImageURL           string    `gorm:"size:500" json:"imageUrl"`
	OptionsJSON        string    `gorm:"type:text" json:"optionsJson"`
	CorrectAnswer      string    `gorm:"size:500" json:"correctAnswer,omitempty"`
	CorrectAnswersJSON string    `gorm:"type:text" json:"correctAnswersJson,omitempty"`
	Points             int       `gorm:"default:1" json:"points"`
	SortOrder          int       `gorm:"default:1" json:"sortOrder"`
	CreatedAt          time.Time `json:"createdAt"`
	UpdatedAt          time.Time `json:"updatedAt"`
}

type TestAttempt struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	TestID        uint      `gorm:"index;not null" json:"testId"`
	Test          Test      `gorm:"foreignKey:TestID" json:"test,omitempty"`
	StudentID     uint      `gorm:"index;not null" json:"studentId"`
	Student       User      `gorm:"foreignKey:StudentID" json:"student,omitempty"`
	Answers       string    `gorm:"type:text" json:"answers"`
	AttemptNumber int       `gorm:"default:1" json:"attemptNumber"`
	Score         int       `gorm:"default:0" json:"score"`
	AutoScore     int       `gorm:"default:0" json:"autoScore"`
	MaxScore      int       `gorm:"default:0" json:"maxScore"`
	Status        string    `gorm:"size:40;default:completed" json:"status"`
	Comment       string    `gorm:"type:text" json:"comment"`
	ManualGrades  string    `gorm:"type:text" json:"manualGrades"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

type Grade struct {
	ID            uint       `gorm:"primaryKey" json:"id"`
	UserID        uint       `gorm:"index;not null" json:"userId"`
	User          User       `gorm:"foreignKey:UserID" json:"user,omitempty"`
	CourseID      uint       `gorm:"index;not null" json:"courseId"`
	Course        Course     `gorm:"foreignKey:CourseID" json:"course,omitempty"`
	Kind          string     `gorm:"size:40;index;not null" json:"kind"`
	AssignmentID  *uint      `gorm:"index" json:"assignmentId"`
	Assignment    Assignment `gorm:"foreignKey:AssignmentID" json:"assignment,omitempty"`
	TestID        *uint      `gorm:"index" json:"testId"`
	Test          Test       `gorm:"foreignKey:TestID" json:"test,omitempty"`
	TestAttemptID *uint      `gorm:"index" json:"testAttemptId"`
	Value         int        `gorm:"not null" json:"value"`
	MaxValue      int        `gorm:"default:5" json:"maxValue"`
	Comment       string     `gorm:"type:text" json:"comment"`
	CreatedAt     time.Time  `json:"createdAt"`
	UpdatedAt     time.Time  `json:"updatedAt"`
}

type News struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Title       string    `gorm:"size:220;not null" json:"title"`
	Excerpt     string    `gorm:"size:400" json:"excerpt"`
	Content     string    `gorm:"type:text" json:"content"`
	ImageURL    string    `gorm:"size:500" json:"imageUrl"`
	PublishedAt time.Time `json:"publishedAt"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type TeacherProfile struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"size:160;not null" json:"name"`
	Position  string    `gorm:"size:180" json:"position"`
	Email     string    `gorm:"size:160" json:"email"`
	ImageURL  string    `gorm:"size:500" json:"imageUrl"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type Application struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	FirstName string    `gorm:"size:80;not null" json:"firstName"`
	LastName  string    `gorm:"size:80;not null" json:"lastName"`
	Phone     string    `gorm:"size:40;not null" json:"phone"`
	Direction string    `gorm:"size:260" json:"direction"`
	Status    string    `gorm:"size:40;default:new" json:"status"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type ScheduleItem struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	GroupName   string    `gorm:"size:80;index" json:"groupName"`
	CourseID    uint      `gorm:"index" json:"courseId"`
	CourseTitle string    `gorm:"size:180;not null" json:"courseTitle"`
	TeacherName string    `gorm:"size:160" json:"teacherName"`
	Room        string    `gorm:"size:80" json:"room"`
	StartsAt    time.Time `gorm:"index" json:"startsAt"`
	EndsAt      time.Time `json:"endsAt"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}
