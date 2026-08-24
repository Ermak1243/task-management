package integration

import (
	"context"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	"task-management/internal/config"
	mysqlpkg "task-management/internal/database/mysql"
	"task-management/internal/models"
	"task-management/internal/repository"
)

// TestStatsRepositoryTeamStats проверяет корректность SQL-запроса в
// internal/repository/stats.go против реальной MySQL. Бизнес-правило доступа
// (IsAdminOrOwner) здесь не тестируется — оно покрыто юнит-тестом в
// internal/service/stats_test.go.
func TestStatsRepositoryTeamStats(t *testing.T) {
	cfg, err := config.New("../../configs/.env.example")
	if err != nil {
		t.Skip("Skipping integration test: no config")
	}

	dbConn := mysqlpkg.New(cfg.MySQL)
	defer dbConn.Close()
	db := dbConn.SQLX()

	clearTables(t, db)

	user1 := createUser(t, db, "user1@test.com", "password", "User One")
	user2 := createUser(t, db, "user2@test.com", "password", "User Two")
	user3 := createUser(t, db, "user3@test.com", "password", "User Three")

	team := createTeam(t, db, "Test Team", user1.ID)
	addMember(t, db, team.ID, user2.ID, "member")
	addMember(t, db, team.ID, user3.ID, "member")

	now := time.Now()
	createTask(t, db, team.ID, "Task 1", "todo", user1.ID, &user2.ID, now.Add(-1*time.Hour), nil)
	createTask(t, db, team.ID, "Task 2", "in_progress", user1.ID, &user3.ID, now.Add(-2*time.Hour), nil)
	createTask(t, db, team.ID, "Task 3", "done", user2.ID, &user1.ID, now.Add(-3*time.Hour), timePtr(now.Add(-1*time.Hour)))
	createTask(t, db, team.ID, "Task 4", "done", user3.ID, &user2.ID, now.Add(-4*time.Hour), timePtr(now.Add(-30*time.Minute)))
	createTask(t, db, team.ID, "Task 5", "done", user1.ID, &user3.ID, now.Add(-5*time.Hour), timePtr(now.Add(-2*time.Hour)))

	addComment(t, db, 1, user1.ID, "comment 1")
	addComment(t, db, 1, user2.ID, "comment 2")
	addComment(t, db, 2, user3.ID, "comment 3")

	stats, err := repository.NewStats(db).TeamStats(context.Background(), team.ID)
	require.NoError(t, err)

	assert.Equal(t, 1, stats.Todo)
	assert.Equal(t, 1, stats.InProgress)
	assert.Equal(t, 3, stats.Done)
	assert.Greater(t, stats.AvgCloseTime, 0.0)
	assert.Len(t, stats.TopAssignees, 3)
	assert.Equal(t, 3, stats.TotalComments)
}

func clearTables(t *testing.T, db *sqlx.DB) {
	tables := []string{"task_comments", "task_history", "tasks", "team_members", "teams", "users"}
	for _, table := range tables {
		_, err := db.Exec("DELETE FROM " + table)
		require.NoError(t, err)
	}
}

func createUser(t *testing.T, db *sqlx.DB, email, password, name string) *models.User {
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	u := &models.User{
		Email:        email,
		PasswordHash: string(hash),
		Name:         name,
		CreatedAt:    time.Now(),
	}
	res, err := db.Exec(
		"INSERT INTO users (email, password_hash, name, created_at) VALUES (?, ?, ?, ?)",
		u.Email, u.PasswordHash, u.Name, u.CreatedAt,
	)
	require.NoError(t, err)
	id, err := res.LastInsertId()
	require.NoError(t, err)
	u.ID = id
	return u
}

func createTeam(t *testing.T, db *sqlx.DB, name string, createdBy int64) *models.Team {
	team := &models.Team{Name: name, CreatedBy: createdBy, CreatedAt: time.Now()}
	res, err := db.Exec(
		"INSERT INTO teams (name, created_by, created_at) VALUES (?, ?, ?)",
		team.Name, team.CreatedBy, team.CreatedAt,
	)
	require.NoError(t, err)
	id, err := res.LastInsertId()
	require.NoError(t, err)
	team.ID = id
	return team
}

func addMember(t *testing.T, db *sqlx.DB, teamID, userID int64, role string) {
	_, err := db.Exec(
		"INSERT INTO team_members (team_id, user_id, role) VALUES (?, ?, ?)",
		teamID, userID, role,
	)
	require.NoError(t, err)
}

func createTask(t *testing.T, db *sqlx.DB, teamID int64, title, status string, createdBy int64, assigneeID *int64, createdAt time.Time, closedAt *time.Time) {
	_, err := db.Exec(
		`INSERT INTO tasks (team_id, title, description, status, created_by, assignee_id, created_at, updated_at, closed_at, version)
         VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		teamID, title, "", status, createdBy, assigneeID, createdAt, time.Now(), closedAt, 1,
	)
	require.NoError(t, err)
}

func timePtr(t time.Time) *time.Time {
	return &t
}

func addComment(t *testing.T, db *sqlx.DB, taskID, userID int64, content string) {
	_, err := db.Exec(
		"INSERT INTO task_comments (task_id, user_id, content, created_at) VALUES (?, ?, ?, ?)",
		taskID, userID, content, time.Now(),
	)
	require.NoError(t, err)
}