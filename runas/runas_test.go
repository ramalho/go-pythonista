package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"
)

const linhaLetraA = "0041;LATIN CAPITAL LETTER A;Lu;0;L;;;;;N;;;;0061;"

const linhas3Da43 = `
003D;EQUALS SIGN;Sm;0;ON;;;;;N;;;;;
003E;GREATER-THAN SIGN;Sm;0;ON;;;;;Y;;;;;
003F;QUESTION MARK;Po;0;ON;;;;;N;;;;;
0040;COMMERCIAL AT;Po;0;ON;;;;;N;;;;;
0041;LATIN CAPITAL LETTER A;Lu;0;L;;;;;N;;;;0061;
0042;LATIN CAPITAL LETTER B;Lu;0;L;;;;;N;;;;0062;
0043;LATIN CAPITAL LETTER C;Lu;0;L;;;;;N;;;;0063;
`

func TestAnalisarLinha(t *testing.T) {
	runa, nome, palavras := AnalisarLinha(linhaLetraA)
	if runa != 'A' {
		t.Errorf("Esperado: 'A'; recebido: %q", runa)
	}
	const nomeA = "LATIN CAPITAL LETTER A"
	if nome != nomeA {
		t.Errorf("Esperado: %q; recebido: %q", nomeA, nome)
	}
	palavrasA := []string{"LATIN", "CAPITAL", "LETTER", "A"}
	if !reflect.DeepEqual(palavras, palavrasA) {
		t.Errorf("\n\tEsperado: %q\n\trecebido: %q", palavrasA, palavras)
	}
}

func TestAnalisarLinhaComHífenECampo10(t *testing.T) {
	var casos = []struct {
		linha    string
		runa     rune
		nome     string
		palavras []string
	}{
		{"0021;EXCLAMATION MARK;Po;0;ON;;;;;N;;;;;",
			'!', "EXCLAMATION MARK", []string{"EXCLAMATION", "MARK"}},
		{"002D;HYPHEN-MINUS;Pd;0;ES;;;;;N;;;;;",
			'-', "HYPHEN-MINUS", []string{"HYPHEN", "MINUS"}},
		{"0027;APOSTROPHE;Po;0;ON;;;;;N;APOSTROPHE-QUOTE;;;",
			'\'', "APOSTROPHE (APOSTROPHE-QUOTE)", []string{"APOSTROPHE", "QUOTE"}},
	}
	for _, caso := range casos {
		runa, nome, palavras := AnalisarLinha(caso.linha)
		if runa != caso.runa || nome != caso.nome ||
			!reflect.DeepEqual(palavras, caso.palavras) {
			t.Errorf("\nAnalisarLinha(%q)\n-> (%q, %q, %q)",
				caso.linha, runa, nome, palavras)
		}
	}
}

func TestContém(t *testing.T) {
	casos := []struct {
		fatia     []string
		procurado string
		esperado  bool
	}{
		{[]string{"A", "B"}, "B", true},
		{[]string{}, "A", false},
		{[]string{"A", "B"}, "Z", false},
	}
	for _, caso := range casos {
		recebido := contém(caso.fatia, caso.procurado)
		if recebido != caso.esperado {
			t.Errorf("contém(%#v, %#v) esperado: %v; recebido: %v",
				caso.fatia, caso.procurado, caso.esperado, recebido)
		}
	}
}

func TestContémTodas(t *testing.T) {
	casos := []struct {
		fatia      []string
		procurados []string
		esperado   bool
	}{
		{[]string{"A", "B"}, []string{"B"}, true},
		{[]string{}, []string{"A"}, false},
		{[]string{"A"}, []string{}, true},
		{[]string{"A", "B"}, []string{"Z"}, false},
		{[]string{"A", "B", "C"}, []string{"A", "C"}, true},
		{[]string{"A", "B", "C"}, []string{"A", "Z"}, false},
		{[]string{"A", "B"}, []string{"A", "B", "C"}, false},
	}
	for _, caso := range casos {
		obtido := contémTodas(caso.fatia, caso.procurados)
		if obtido != caso.esperado {
			t.Errorf("contémTodas(%#v, %#v)\nesperado: %v; recebido: %v",
				caso.fatia, caso.procurados, caso.esperado, obtido)
		}
	}
}

func TestSeparar(t *testing.T) {
	casos := []struct {
		texto    string
		esperado []string
	}{
		{"A", []string{"A"}},
		{"A B", []string{"A", "B"}},
		{"A B-C", []string{"A", "B", "C"}},
	}
	for _, caso := range casos {
		obtido := separar(caso.texto)
		if !reflect.DeepEqual(obtido, caso.esperado) {
			t.Errorf("separar(%q)\nesperado: %#v; recebido: %#v",
				caso.texto, caso.esperado, obtido)
		}
	}
}

func ExampleListar() {
	texto := strings.NewReader(linhas3Da43)
	Listar(texto, "MARK")
	// Output: U+003F	?	QUESTION MARK
}

func ExampleListar_doisResultados() {
	texto := strings.NewReader(linhas3Da43)
	Listar(texto, "SIGN")
	// Output:
	// U+003D	=	EQUALS SIGN
	// U+003E	>	GREATER-THAN SIGN
}

func ExampleListar_duasPalavras() {
	texto := strings.NewReader(linhas3Da43)
	Listar(texto, "CAPITAL LATIN")
	// Output:
	// U+0041	A	LATIN CAPITAL LETTER A
	// U+0042	B	LATIN CAPITAL LETTER B
	// U+0043	C	LATIN CAPITAL LETTER C
}

func Example() {
	argsAntes := os.Args
	defer func() { os.Args = argsAntes }()
	os.Args = []string{"", "cruzeiro"}
	main()
	// Output:
	// U+20A2	₢	CRUZEIRO SIGN
}

func Example_consultaDuasPalavras() {
	argsAntes := os.Args
	defer func() { os.Args = argsAntes }()
	os.Args = []string{"", "cat", "smiling"}
	main()
	// Output:
	// U+1F638	😸	GRINNING CAT FACE WITH SMILING EYES
	// U+1F63A	😺	SMILING CAT FACE WITH OPEN MOUTH
	// U+1F63B	😻	SMILING CAT FACE WITH HEART-SHAPED EYES
}

func Example_consultaComHífenECampo10() {
	argsAntes := os.Args
	defer func() { os.Args = argsAntes }()
	os.Args = []string{"", "quote"}
	main()
	// Output:
	// U+0027	'	APOSTROPHE (APOSTROPHE-QUOTE)
	// U+2358	⍘	APL FUNCTIONAL SYMBOL QUOTE UNDERBAR
	// U+235E	⍞	APL FUNCTIONAL SYMBOL QUOTE QUAD
}

func restaurar(nomeVar, valor string, existia bool) {
	if existia {
		os.Setenv(nomeVar, valor)
	} else {
		os.Unsetenv(nomeVar)
	}
}

func TestObterCaminhoUCD_setado(t *testing.T) {
	caminhoAntes, existia := os.LookupEnv("UCD_PATH")
	defer restaurar("UCD_PATH", caminhoAntes, existia)
	caminhoUCD := fmt.Sprintf("./TEST%d-UnicodeData.txt", time.Now().UnixNano())
	os.Setenv("UCD_PATH", caminhoUCD)
	obtido := obterCaminhoUCD()
	if obtido != caminhoUCD {
		t.Errorf("obterCaminhoUCD() [setado]\nesperado: %q; recebido: %q",
			caminhoUCD, obtido)
	}
}

func TestObterCaminhoUCD_default(t *testing.T) {
	caminhoAntes, existia := os.LookupEnv("UCD_PATH")
	defer restaurar("UCD_PATH", caminhoAntes, existia)
	os.Unsetenv("UCD_PATH")
	sufixoCaminhoUCD := "/UnicodeData.txt"
	obtido := obterCaminhoUCD()
	if !strings.HasSuffix(obtido, sufixoCaminhoUCD) {
		t.Errorf("obterCaminhoUCD() [default]\nesperado (sufixo): %q; recebido: %q",
			sufixoCaminhoUCD, obtido)
	}
}

func TestBaixarUCD(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(linhas3Da43))
		}))
	defer srv.Close()

	caminhoUCD := fmt.Sprintf("./TEST%d-UnicodeData.txt", time.Now().UnixNano())
	feito := make(chan bool)
	go baixarUCD(srv.URL, caminhoUCD, feito)
	_ = <-feito
	ucd, err := os.Open(caminhoUCD)
	if os.IsNotExist(err) {
		t.Errorf("baixarUCD não gerou:%v\n%v", caminhoUCD, err)
	}
	ucd.Close()
	os.Remove(caminhoUCD)
}

func TestAbrirUCD_local(t *testing.T) {
	caminhoUCD := obterCaminhoUCD()
	ucd, err := abrirUCD(caminhoUCD)
	if err != nil {
		t.Errorf("AbrirUCD(%q):\n%v", caminhoUCD, err)
	}
	ucd.Close()
}

func TestAbrirUCD_remoto(t *testing.T) {
	if testing.Short() {
		t.Skip("teste ignorado [opção -test.short]")
	}
	caminhoUCD := fmt.Sprintf("./TEST%d-UnicodeData.txt", time.Now().UnixNano())
	ucd, err := abrirUCD(caminhoUCD)
	if err != nil {
		t.Errorf("AbrirUCD(%q):\n%v", caminhoUCD, err)
	}
	ucd.Close()
	os.Remove(caminhoUCD)
}
