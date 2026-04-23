package main

import (
	"fmt"     //importamos format pq se necesita para imprimir en la consola y para las etiquetas en HTML
	"html"    //esta es porque si
	"log"     //añadido porque usas log.Fatalf
	"os"      //importamos os para poder leer el codigo de python y para crear el archivo final con los colores
	"regexp"  //reges es importante porque es el que va analizar cada palabra y ver a que pertenece, permite definir las reglas de busquedas
	"sort"    //sort es para cuando lea nuestro python lo ponga en orden para poder asignarles el color correcto
	"strings" //es para manipular el texto, ayuda a construir el HTML y no gasta mucha memoria
)

type Token struct { //Token es una pieza de codigo detectada por su posicion y lo que es
	start    int    // aqui es el indice de donde empieza en el texto
	end      int    //Aqui es el indice donde termina
	category string //Aqui es para la categoria
}

var colores = map[string]string{ //el color que le estamos asignando  a cada cosa
	"ciclo":       "#1865bd",
	"comentario":  "#2d7931",
	"variable":    "#791cb2",
	"tipo":        "#e85607",
	"funcion":     "#b61b1b",
	"clase":       "#910b52",
	"condicional": "#f2a528",
	"import":      "#067f8a",
}

var prioridad = map[string]int{ //Prioridad es para decidir que color va a ganar si es que dos patrones son iguales
	"comentario": 8, "funcion": 7, "clase": 6, "import": 5, //el comentario que es el numero 8 es el de prioridad mas alta, no se va a pintar nada de un coclor aunque diga def o una variable porque el comentario tiene prioridad
	"condicional": 4, "ciclo": 3, "tipo": 2, "variable": 1,
}

func main() { //aqui lee el archivo, abrimos el .txt con nuestro codigo de Python que es el lenguaje que estamos analizando
	data, err := os.ReadFile("python.txt")
	if err != nil {
		log.Fatalf("Error reading file: %v", err)
	}
	source := string(data) //necesitamos pasar data a string para que regex pueda leerlo

	type Patron struct { //definimos los patrones, los vamos a definir usando regex. Regex se usa para buscar en un codigo un patron, el patron que bsuscas lo deifnimos aca abajo en patron
		cat string         //nombre de la categoria
		re  *regexp.Regexp //la expresion compilada
	}
	patrones := []Patron{
		//?s permite que incluya saltos de linea y ?m hace que detecte el inicio de cada nueva linea
		{"comentario", regexp.MustCompile(`(?s)(""".*?"""|'''.*?'''|#[^\n]*)`)},
		{"clase", regexp.MustCompile(`(?m)^[ \t]*class[ \t]+[A-Za-z_]\w*.*:`)},
		{"funcion", regexp.MustCompile(`(?m)(^[ \t]*def[ \t]+[A-Za-z_]\w*[ \t]*\([^)]*\)[ \t]*:|^[ \t]*return\b[^\n]*)`)},
		{"import", regexp.MustCompile(`(?m)^[ \t]*(import[ \t]+[^\n]+|from[ \t]+[^\n]+)`)},
		{"condicional", regexp.MustCompile(`(?m)^[ \t]*(if|elif|else)(\b[^\n]*)?:`)},
		{"ciclo", regexp.MustCompile(`(?m)^[ \t]*(for|while)\b[^\n]*:`)},
		{"tipo", regexp.MustCompile(`\b(int|float|str|bool|string|list|dict)\b`)},
		{"variable", regexp.MustCompile(`(?m)^[ \t]*([A-Za-z_]\w*)[ \t]*(?::[^=\n]+)?=`)},
	}

	var tokens []Token //Aqui buscamos los tokens, es para que el .html funcione. Buscamos en el texto coincidencias de los patrones.
	for _, p := range patrones {
		for _, m := range p.re.FindAllStringSubmatchIndex(source, -1) {
			s, e := m[0], m[1]
			if p.cat == "variable" && len(m) >= 4 { //este es para las variables que solo le de color a las variables y no al igual =
				s, e = m[2], m[3]
			}
			tokens = append(tokens, Token{s, e, p.cat})
		}
	}

	sort.Slice(tokens, func(i, j int) bool { //Aqui ordenamos los tokens por orden de aparicion, si dos tokens estan en el mismo lugar entonces lo definimos por el que tenga mayor prioridad
		if tokens[i].start != tokens[j].start {
			return tokens[i].start < tokens[j].start
		}
		return prioridad[tokens[i].category] > prioridad[tokens[j].category]
	})
	//Este es para evitar que los colores se encimen
	var resueltos []Token
	cursor := 0
	for _, t := range tokens {
		if t.start >= cursor { //si el token actual empieza despues de donde termino el anterior
			resueltos = append(resueltos, t)
			cursor = t.end //movemos el cursos al final de este token
		}
	}

	var codigo strings.Builder //Aqui ya estamos construyendo el .html, estamos poniendo los colores
	cursor = 0
	for _, t := range resueltos {
		if t.start > cursor {
			codigo.WriteString(html.EscapeString(source[cursor:t.start]))
		} //Aqui añadimos el texto ya con color
		fmt.Fprintf(&codigo, `<span style="color:%s">%s</span>`, colores[t.category], html.EscapeString(source[t.start:t.end]))
		cursor = t.end
	}

	if cursor < len(source) {
		codigo.WriteString(html.EscapeString(source[cursor:]))
	}
	//aqui ponemos la numeracion de cada linea
	var lineas strings.Builder
	for i, l := range strings.Split(codigo.String(), "\n") {
		fmt.Fprintf(&lineas, "<tr><td style='color:#aaa;padding-right:16px;user-select:none;text-align:right'>%d</td><td style='white-space:pre'>%s</td></tr>\n", i+1, l)
	}
	//Esta es nuestra salida, aqui ponemos lo que queremos que tenga el .html
	salida := `<!DOCTYPE html>
<html lang="es">
<head>
  <meta charset="UTF-8">
  <title>Analizador Lexico</title>
  <style>
    body { background: #fff; font-family: monospace; font-size: 14px; padding: 20px; }
    p { margin-bottom: 12px; }
  </style>
</head>
<body>
  <p>
    <b>Leyenda:</b>
    <span style="color:#1865bd">ciclo</span> &nbsp;
    <span style="color:#2d7931">comentario</span> &nbsp;
    <span style="color:#791cb2">variable</span> &nbsp;
    <span style="color:#e85607">tipo</span> &nbsp;
    <span style="color:#b61b1b">funcion</span> &nbsp;
    <span style="color:#910b52">clase</span> &nbsp;
    <span style="color:#f2a528">condicional</span> &nbsp;
    <span style="color:#067f8a">import</span>
  </p>
  <table style="border-collapse:collapse"><tbody>` + lineas.String() + `</tbody></table>
</body>
</html>`

	os.WriteFile("out.html", []byte(salida), 0644)
	fmt.Println("Se ha generado el out.html")
}
