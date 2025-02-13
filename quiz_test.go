package main

import (
	"os"
	"strings"
	"testing"
)

func TestLoadQuiz(t *testing.T) {
	tempFile, err := os.CreateTemp("", "quiz_test.csv")
	if err != nil {
		t.Fatalf("Impossible de créer un fichier temporaire : %v", err)
	}
	defer os.Remove(tempFile.Name())

	content := "5+5,10\n2+2,4\nQuelle est la capitale de la France?,Paris\n"
	tempFile.WriteString(content)
	tempFile.Close()

	quiz := loadQuiz(tempFile.Name())

	expectedQuestions := 3
	if len(quiz.questions) != expectedQuestions {
		t.Errorf("Nombre de questions attendu : %d, obtenu : %d", expectedQuestions, len(quiz.questions))
	}
}

func TestShuffleQuestions(t *testing.T) {
	quiz := &quiz{
		questions: []question{
			{"Q1", "A1"},
			{"Q2", "A2"},
			{"Q3", "A3"},
		},
	}

	initialOrder := make([]question, len(quiz.questions))
	copy(initialOrder, quiz.questions)

	quiz.shuffleQuestions()

	shuffled := false
	for i := range quiz.questions {
		if quiz.questions[i] != initialOrder[i] {
			shuffled = true
			break
		}
	}

	if !shuffled {
		t.Errorf("Les questions n'ont pas été mélangées correctement")
	}
}

func TestSaveResults(t *testing.T) {
	outputPath := "test_results.csv"
	defer os.Remove(outputPath)

	quiz := &quiz{
		answered:          2,
		answeredCorrectly: 1,
		questions: []question{
			{"Q1", "A1"},
			{"Q2", "A2"},
		},
		userAnswers: map[string]string{
			"Q1": "A1",
			"Q2": "wrong",
		},
	}

	originalQuestions := []question{
		{"Q1", "A1"},
		{"Q2", "A2"},
	}

	saveResults(outputPath, "TestUser", quiz, originalQuestions)

	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Erreur lors de la lecture du fichier : %v", err)
	}

	expectedContent := "TestUser,A1,wrong,2,1,50.00%"
	if !strings.Contains(string(data), expectedContent) {
		t.Errorf("Contenu du fichier incorrect :\n%s", string(data))
	}
}
