package main

import (
	"bufio"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
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
	userAnswers       []string
}

// Chargement du quiz depuis un fichier CSV
func loadQuiz(filePath string) *quiz {
	csvFile, err := os.Open(filePath)
	fatalError("Erreur lors de l'ouverture du fichier CSV", err)
	defer csvFile.Close()
	reader := csv.NewReader(bufio.NewReader(csvFile))
	var quiz quiz
	for {
		line, err := reader.Read()
		if err == io.EOF {
			break
		}
		fatalError("Erreur lors de l'analyse du CSV", err)
		question := question{line[0], line[1]}
		quiz.questions = append(quiz.questions, question)
	}
	return &quiz
}

// Exécution du quiz
func (quiz *quiz) run() {
	fmt.Println("Merci de répondre par un chiffre, un nombre ou un mot sans majuscule.")
	fmt.Printf("Attention, le quiz est limité à %d secondes.\n", *timeLimit)
	timer := time.NewTimer(time.Duration(*timeLimit) * time.Second)
quizLoop:
	for _, question := range quiz.questions {
		fmt.Println(question.question)
		answerCh := make(chan string)
		go func() {
			scanner.Scan()
			answer := scanner.Text()
			answerCh <- answer
		}()
		select {
		case <-timer.C:
			break quizLoop
		case answer := <-answerCh:
			quiz.userAnswers = append(quiz.userAnswers, answer)
			if answer == question.answer {
				quiz.answeredCorrectly++
			}
			quiz.answered++
		}
	}
	for len(quiz.userAnswers) < len(quiz.questions) {
		quiz.userAnswers = append(quiz.userAnswers, "")
	}
	return
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
// Si le fichier n'existe pas encore, créé un en-tête de correction
func saveResultsHeader(outputPath string, questions []question) {
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		csvFile, err := os.Create(outputPath)
		fatalError("Erreur lors de la création du fichier CSV des résultats", err)
		defer csvFile.Close()
		writer := csv.NewWriter(csvFile)
		defer writer.Flush()

		// Écrit la ligne de correction avec les réponses correctes
		correctionRow := []string{"Correction"}
		for _, q := range questions {
			correctionRow = append(correctionRow, q.answer)
		}
		correctionRow = append(correctionRow, "Nombre de questions répondues", "Nombre de réponses correctes")
		writer.Write(correctionRow)
	}
}

func saveResults(outputPath string, userName string, quiz *quiz) {
	csvFile, err := os.OpenFile(outputPath, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	fatalError("Erreur lors de l'ouverture du fichier CSV pour écrire les résultats", err)
	defer csvFile.Close()
	writer := csv.NewWriter(csvFile)
	defer writer.Flush()

	// Écrit les réponses de l'utilisateur
	userRow := []string{userName}
	userRow = append(userRow, quiz.userAnswers...)
	// Ajoute les deux colonnes supplémentaires à la fin
	userRow = append(userRow, fmt.Sprintf("%d", quiz.answered), fmt.Sprintf("%d", quiz.answeredCorrectly))
	writer.Write(userRow)
}

var (
	scanner     = bufio.NewScanner(os.Stdin)
	filePathPtr = flag.String("file", "./problemes.csv", "Fichier contenant les questions du quiz")
	timeLimit   = flag.Int64("time-limit", 30, "Temps limite pour répondre aux questions")
)

func main() {
	outputPath := "resultats_quiz.csv"

	// Charger le quiz
	quizData := loadQuiz(*filePathPtr)

	// Initialiser le fichier CSV avec l'en-tête si nécessaire
	saveResultsHeader(outputPath, quizData.questions)

	for {
		// Demande le nom de l'utilisateur
		fmt.Print("Entrez votre nom : ")
		scanner.Scan()
		userName := scanner.Text()

		// Préparer un nouveau quiz
		quiz := loadQuiz(*filePathPtr)

		// Exécuter le quiz
		quiz.run()

		// Afficher le resultat
		quiz.report(userName)

		// Enregistrer les résultats dans le fichier CSV
		saveResults(outputPath, userName, quiz)
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
