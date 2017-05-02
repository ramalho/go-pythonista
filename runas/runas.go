package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/user"
	"strconv"
	"strings"
	"time"
)

// URLUCD é a URL canônica do arquivo UnicodeData.txt mais atual
const URLUCD = "http://www.unicode.org/Public/UNIDATA/UnicodeData.txt"

func separar(s string) []string {
	s = strings.Replace(s, "-", " ", -1)
	return strings.Split(s, " ")
}

// AnalisarLinha devolve a runa, o nome e uma fatia de palavras
// que ocorrem no campo nome de uma linha do UnicodeData.txt
func AnalisarLinha(linha string) (rune, string, []string) {
	campos := strings.Split(linha, ";")
	código, _ := strconv.ParseInt(campos[0], 16, 32)
	nome := campos[1]
	palavras := separar(campos[1])
	if campos[10] != "" {
		nome += fmt.Sprintf(" (%s)", campos[10])
		for _, palavra := range separar(campos[10]) {
			if !contém(palavras, palavra) {
				palavras = append(palavras, palavra)
			}
		}
	}
	return rune(código), nome, palavras
}

func contém(fatia []string, procurado string) bool {
	for _, item := range fatia {
		if item == procurado {
			return true
		}
	}
	return false
}

func contémTodas(fatia []string, procurados []string) bool {
	for _, procurado := range procurados {
		if !contém(fatia, procurado) {
			return false
		}
	}
	return true
}

// Listar exibe na saída padrão o código, a runa e o nome dos
// caracteres Unicode onde ocorrem as palavras da consulta.
func Listar(texto io.Reader, consulta string) {
	termos := separar(consulta)
	varredor := bufio.NewScanner(texto)
	for varredor.Scan() {
		linha := varredor.Text()
		if strings.TrimSpace(linha) == "" {
			continue
		}
		runa, nome, palavrasNome := AnalisarLinha(linha)
		if contémTodas(palavrasNome, termos) {
			fmt.Printf("U+%04X\t%[1]c\t%s\n", runa, nome)
		}
	}
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func obterCaminhoUCD() string {
	caminhoUCD := os.Getenv("UCD_PATH")
	if caminhoUCD == "" {
		usuário, err := user.Current()
		check(err)
		caminhoUCD = usuário.HomeDir + "/UnicodeData.txt"
	}
	return caminhoUCD
}

func progresso(feito <-chan bool) {
	for {
		select {
		case <-feito:
			fmt.Println()
			return
		default:
			fmt.Print(".")
			time.Sleep(150 * time.Millisecond)
		}
	}
}

func baixarUCD(url, caminho string, feito chan<- bool) {
	resposta, err := http.Get(url)
	check(err)
	defer resposta.Body.Close()
	arquivo, err := os.Create(caminho)
	check(err)
	defer arquivo.Close()
	_, err = io.Copy(arquivo, resposta.Body)
	check(err)
	feito <- true
}

func abrirUCD(caminho string) (*os.File, error) {
	ucd, err := os.Open(caminho)
	if os.IsNotExist(err) {
		fmt.Printf("%s não encontrado\nbaixando %s\n", caminho, URLUCD)
		feito := make(chan bool)
		go baixarUCD(URLUCD, caminho, feito)
		progresso(feito)
		ucd, err = os.Open(caminho)
	}
	return ucd, err
}

func main() {
	ucd, err := abrirUCD(obterCaminhoUCD())
	if err != nil {
		log.Fatal(err.Error())
	}
	defer func() { ucd.Close() }()
	consulta := strings.Join(os.Args[1:], " ")
	Listar(ucd, strings.ToUpper(consulta))
}
