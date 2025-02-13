package main

import (
	"bufio"
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"
)

// Gestion des erreurs
func fatalError(message string, err error) {
	if err != nil {
		log.Fatalln(message, ":", err)
	}
}

// Une question est définie par la question et sa réponse
type question struct {
	question string
	answer   string
}

// Le quiz est défini par :
// - le nombre de questions répondues par l'utilisateur
// - le nombre de réponses correctes
// - la liste de questions
// - les réponses de l'utilisateur
type quiz struct {
	answered          int
	answeredCorrectly int
	questions         []question
	userAnswers       map[string]string
}

// Chargement du quiz depuis un fichier CSV
func loadQuiz(filePath string) *quiz {
	csvFile, err := os.Open(filePath)
	fatalError("Erreur lors de l'ouverture du fichier CSV", err)
	defer csvFile.Close()

	reader := csv.NewReader(csvFile)
	var quiz quiz
	quiz.userAnswers = make(map[string]string)

	lines, err := reader.ReadAll()
	fatalError("Erreur lors de l'analyse du CSV", err)

	for _, line := range lines {
		if len(line) < 2 {
			continue
		}
		q := question{line[0], line[1]}
		quiz.questions = append(quiz.questions, q)
	}

	return &quiz
}

// Fonction pour mélanger les questions
func (quiz *quiz) shuffleQuestions() {
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(quiz.questions), func(i, j int) {
		quiz.questions[i], quiz.questions[j] = quiz.questions[j], quiz.questions[i]
	})
}

// Exécution du quiz
func (quiz *quiz) run() {
	fmt.Println("Merci de répondre par un chiffre, un nombre ou un mot sans majuscule.")
	fmt.Printf("Attention, le quiz est limité à %d secondes.\n", *timeLimit)
	timer := time.NewTimer(time.Duration(*timeLimit) * time.Second)
quizLoop:
	for _, q := range quiz.questions {
		fmt.Println(q.question)
		answerCh := make(chan string)
		go func() {
			scanner.Scan()
			answerCh <- scanner.Text()
		}()
		select {
		case <-timer.C:
			break quizLoop
		case answer := <-answerCh:
			quiz.userAnswers[q.question] = answer
			if answer == q.answer {
				quiz.answeredCorrectly++
			}
			quiz.answered++
		}
	}
}

// Affichage du resultat de l'utilisateur
func (quiz *quiz) report(userName string) {
	fmt.Printf(
		"%s, le quiz est terminé ! Vous avez répondu à %v questions sur %v et %v sont correctes.\n",
		userName,
		quiz.answered,
		len(quiz.questions),
		quiz.answeredCorrectly,
	)
}

// Sauvegarde des résultats dans un fichier CSV
func saveResults(outputPath string, userName string, quiz *quiz, originalQuestions []question) {
	file, err := os.OpenFile(outputPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	fatalError("Erreur lors de l'ouverture du fichier de résultats", err)
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Réorganiser les réponses de l'utilisateur selon l'ordre original des questions
	orderedAnswers := make([]string, len(originalQuestions))
	for i, originalQuestion := range originalQuestions {
		orderedAnswers[i] = quiz.userAnswers[originalQuestion.question]
	}

	// Enregistrement des résultats
	result := []string{userName}
	result = append(result, orderedAnswers...)
	result = append(result, fmt.Sprintf("%d", quiz.answered), fmt.Sprintf("%d", quiz.answeredCorrectly), fmt.Sprintf("%.2f%%", float64(quiz.answeredCorrectly)/float64(len(originalQuestions))*100))
	writer.Write(result)
}

var (
	scanner     = bufio.NewScanner(os.Stdin)
	filePathPtr = flag.String("file", "./problemes.csv", "Fichier contenant les questions du quiz")
	timeLimit   = flag.Int64("time-limit", 30, "Temps limite pour répondre aux questions")
)

func main() {
	flag.Parse()
	outputPath := "resultats_quiz.csv"

	// Charger le quiz
	quizData := loadQuiz(*filePathPtr)

	// Sauvegarder une copie des questions d'origine
	originalQuestions := make([]question, len(quizData.questions))
	copy(originalQuestions, quizData.questions)

	for {
		// Demande le nom de l'utilisateur
		fmt.Print("Entrez votre nom : ")
		scanner.Scan()
		userName := scanner.Text()

		// Demande si l'utilisateur souhaite mélanger les questions
		fmt.Print("Voulez-vous mélanger les questions ? (oui/non) : ")
		scanner.Scan()
		shuffleChoice := scanner.Text()

		// Préparer un nouveau quiz
		quiz := loadQuiz(*filePathPtr)

		// Mélanger si option choisie
		if shuffleChoice == "oui" {
			quiz.shuffleQuestions()
		}

		// Exécuter le quiz
		quiz.run()

		// Afficher le resultat
		quiz.report(userName)

		// Enregistrer les résultats dans le fichier CSV
		saveResults(outputPath, userName, quiz, originalQuestions)
		fmt.Printf("Les résultats ont été sauvegardés dans le fichier : %s\n", outputPath)

		// Demander si un autre utilisateur souhaite participer
		fmt.Print("Une autre personne veut-elle faire le quiz ? (oui/non) : ")
		scanner.Scan()
		another := scanner.Text()
		if another != "oui" {
			break
		}
	}

	fmt.Println("Merci d'avoir utilisé le quiz !")
}
