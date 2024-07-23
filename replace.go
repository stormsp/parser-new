package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"regexp"
	"strings"
	"time"
)

type Rep struct {
	sys_num int
	SYS_NUM int
	Value   float32
}

var aout [100]float32
var dout [100]float32

// Время запуска программы
var startTime = time.Now()

var Reps map[string]Rep

func translate_for_to_go(code string) string {

	//добавляем func main() после последнего endfunc
	var newCode string
	// Найдем последнее вхождение "endfunc"
	newCode = code
	lastEndfuncIndex := strings.LastIndex(newCode, "endfunc")
	if lastEndfuncIndex == -1 {
		log.Fatal("Не удалось найти endfunc в коде")
	}

	// Код, который вы хотите добавить после последнего endfunc
	additionalCode := `

func mainOutput()
`
	// Вставим код после последнего endfunc
	code = newCode[:lastEndfuncIndex+7] + additionalCode + newCode[lastEndfuncIndex+7:]

	// Перевод символов

	code = strings.ReplaceAll(code, "&", " && ")
	code = strings.ReplaceAll(code, "|", " || ")
	//code = ReplaceAllStringRegexp(code, `(.+);.*`, "$1")
	code = strings.ReplaceAll(code, ";", "//")
	code = strings.ReplaceAll(code, "#", "//")
	code = ReplaceAllStringRegexp(code, `#\[(.*?)\]`, "$1")
	code = ReplaceAllStringRegexp(code, `(?i)\s*end\w*`, "\n}")

	// изменение func и добавление any после каждой переменной
	code = ReplaceAllStringRegexpFunc(code, `(?i)(func[ \t]+)(\w+\s*\(\s*[^)]*\s*\))\s*`, func(match string) string {
		// Извлекаем имя функции и параметры из совпадения
		reg := regexp.MustCompile(`(?i)(func[ \t]+)(\w+)\s*\(([^)]*)\)`)
		matches := reg.FindStringSubmatch(match)

		// Если имя функции "main", пропускаем изменения
		if strings.EqualFold(matches[2], "main") {
			return "func main()) {\n"
		}

		// Извлекаем параметры из совпадения
		paramsStart := len("func")
		paramsEnd := len(match) - 1
		params := match[paramsStart:paramsEnd]

		// Разбиваем параметры по запятой и добавляем " any" после каждой переменной
		paramArray := regexp.MustCompile(`\s*,\s*`).Split(params, -1)
		for i, param := range paramArray {
			param = strings.TrimSpace(param)
			// Добавляем проверку на "id" или "ID"
			if strings.EqualFold(param, "") {
				paramArray[i] = param
			} else if strings.EqualFold(param, "id)") || strings.EqualFold(param, "timeout") || strings.EqualFold(param, "id") {
				paramArray[i] = param + " int"
			} else {
				paramArray[i] = param + " float32"
			}
			//paramArray[i] = strings.TrimSpace(param) + " any"
		}

		// Удаляем ")" перед последним "any"
		if len(paramArray) > 0 {
			lastParamIndex := len(paramArray) - 1
			paramArray[lastParamIndex] = strings.TrimSuffix(paramArray[lastParamIndex], ")") + ")"
		}
		// Проверяем имя функции и определяем тип возвращаемого значения
		var returnType string
		if strings.EqualFold(matches[2], "checkPrecondSt") || strings.EqualFold(matches[2], "checkPrecondBt") || strings.EqualFold(matches[2], "setwex") {
			returnType = " bool"
		} else {
			if strings.EqualFold(matches[2], "mainOutput") {
				returnType = ""
			} else {
				returnType = " float32"
			}
		}

		// Собираем обновленные параметры и тип возвращаемого значения
		updatedParams := strings.Join(paramArray, ", ")

		// Возвращаем обновленный код
		return "func " + updatedParams + returnType + " {\n\t"
	})
	code = ReplaceAllStringRegexp(code, `func\s+(\w+)\s*\(([^)]*)\)`, `func $1($2`)

	// Дополнительная замена для mainOutput(any)
	code = ReplaceAllStringRegexp(code, `mainOutput\s*\(\s*float32\s*\)`, `mainOutput()`)
	// Перевод функций математики
	code = ReplaceAllStringRegexp(code, `(?i)abs\(([^)]+)\)`, `math.Abs($1)`)
	code = ReplaceAllStringRegexp(code, `(?i)acos\(([^)]+)\)`, `math.Acos($1)`)
	code = ReplaceAllStringRegexp(code, `(?i)asin\(([^)]+)\)`, `math.Asin($1)`)
	code = ReplaceAllStringRegexp(code, `(?i)atan\(([^)]+)\)`, `math.Atan($1)`)
	code = ReplaceAllStringRegexp(code, `(?i)cos\(([^)]+)\)`, `math.Cos($1)`)
	code = ReplaceAllStringRegexp(code, `(?i)sin\(([^)]+)\)`, `math.Sin($1)`)
	code = ReplaceAllStringRegexp(code, `(?i)tan\(([^)]+)\)`, `math.Tan($1)`)
	code = ReplaceAllStringRegexp(code, `(?i)exp\(([^)]+)\)`, `math.Exp($1)`)
	code = ReplaceAllStringRegexp(code, `(?i)ln\(([^)]+)\)`, `math.Log($1)`)
	code = ReplaceAllStringRegexp(code, `(?i)log\(([^)]+)\)`, `math.Log10($1)`)
	code = ReplaceAllStringRegexp(code, `(?i)sqrt\(([^)]+)\)`, `math.Sqrt($1)`)
	code = ReplaceAllStringRegexp(code, `(?i)sign\(([^)]+)\)`, `math.Sign($1)`)
	code = ReplaceAllStringRegexp(code, `(?i)sgn\(([^)]+)\)`, `(0 if $1 == 0 else -1 if $1 < 0 else 1)`)
	code = ReplaceAllStringRegexp(code, `(?i)pow\(([^,]+),([^)]+)\)`, `math.Pow($1, $2)`)
	code = ReplaceAllStringRegexp(code, `(?i)rand\(\)`, `random.Float64()`)
	code = ReplaceAllStringRegexp(code, `(?i)min\(([^,]+),([^)]+)\)`, `math.Min($1, $2)`)
	code = ReplaceAllStringRegexp(code, `(?i)max\(([^,]+),([^)]+)\)`, `math.Max($1, $2)`)
	code = ReplaceAllStringRegexp(code, `(?i)int\(([^)]+)\)`, `int($1)`)
	code = ReplaceAllStringRegexp(code, `(?i)restdiv\(([^,]+),([^)]+)\)`, `$1 % $2`)
	// Для функций nmin и nmax нужно будет написать функцию внутри Go
	code = ReplaceAllStringRegexp(code, `(?i)nmin\(([^)]+)\)`, `NMIN($1)`)
	code = ReplaceAllStringRegexp(code, `(?i)nmax\(([^)]+)\)`, `NMAX($1)`)

	// Логические операции
	// Dost TRUE FALSE нужно будет написать функции внутри GO, потому что такой альтернативы нет

	code = ReplaceAllStringRegexp(code, `(?i)\b(eq)\s*\(([^()]+(?:\{[^}]+\})?),([^)]+)\)`, `(($2) == ($3))`)
	code = ReplaceAllStringRegexp(code, `(?i)\b(eq)\(([^,]+(?:\([^)]+\))?),([^)]+)\)`, `(($2) == ($3))`)

	code = ReplaceAllStringRegexp(code, `(?i)\bne\(([^,]+?(?:\([^)]+\))?),([^)]+)\)`, `($1) != ($2)`)
	code = ReplaceAllStringRegexp(code, `(?i)\b(ge)\(([^,]+(?:\([^)]+\))?),([^)]+)\)`, `(($2) >= ($3))`)
	code = ReplaceAllStringRegexp(code, `(?i)\b(lt)\(([^,]+(?:\([^)]+\))?),([^)]+)\)`, `(($2) < ($3))`)
	code = ReplaceAllStringRegexp(code, `(?i)\b(gt)\(([^,]+(?:\([^)]+\))?),([^)]+)\)`, `(($2) > ($3))`)
	code = ReplaceAllStringRegexp(code, `(?i)\b(le)\(([^,]+(?:\([^)]+\))?),([^)]+)\)`, `(($2) <= ($3))`)
	code = ReplaceAllStringRegexp(code, `(?i)\b(NOT)\(([^)]+)\)`, `^(0xFFFFFFFFFFFFFFFF & $3)`)
	//dost
	code = ReplaceAllStringRegexp(code, `(?i)dost`, "DOST")
	code = ReplaceAllStringRegexp(code, `(?i)\btrue\(([^)]+)\)`, `TRUE($1)`)
	code = ReplaceAllStringRegexp(code, `(?i)\bfalse\(([^)]+)\)`, `FALSE($1)`)

	//конвертируем в инт
	// Работа с битами и байтами
	// Функция BIT
	code = ReplaceAllStringRegexp(code, `bit\(([^,]+),([^)]+)\)`, `$1 & (1 << $2)`)
	// Функция BITS - может потребоваться вспомогательная функция
	code = ReplaceAllStringRegexp(code, `bits\(([^,]+),([^,]+),([^)]+)\)`, `BITS($1, $2, $3)`)
	// Функция BXCHG - может потребоваться вспомогательная функция
	code = ReplaceAllStringRegexp(code, `bxchg\(([^,]+),([^)]+)\)`, `BXCHG($1, $2)`)
	// Функция SETBITS - может потребоваться вспомогательная функция
	code = ReplaceAllStringRegexp(code, `setbits\(([^,]+),([^,]+),([^,]+),([^)]+)\)`, `SETBITS($1, $2, $3, $4)`)

	// Функции над временем
	// Функция TIME
	code = ReplaceAllStringRegexp(code, `time\(\)`, `time.Now().Unix()`)
	// Функция SECOND
	code = ReplaceAllStringRegexp(code, `second\(([^)]+)\)`, `$1.Second()`)
	// Функция MINUTE
	code = ReplaceAllStringRegexp(code, `minute\(([^)]+)\)`, `$1.Minute()`)
	// Функция HOUR
	code = ReplaceAllStringRegexp(code, `hour\(([^)]+)\)`, `$1.Hour()`)
	// Функция MONTHDAY
	code = ReplaceAllStringRegexp(code, `monthday\(([^)]+)\)`, `$1.Day()`)
	// Функция MONTH
	code = ReplaceAllStringRegexp(code, `month\(([^)]+)\)`, `int($1.Month()) - 1`) // В Go месяцы начинаются с 1, а не с 0
	// Функция YEAR
	code = ReplaceAllStringRegexp(code, `year\(([^)]+)\)`, `$1.Year()`)
	// Функция WEEKDAY
	code = ReplaceAllStringRegexp(code, `weekday\(([^)]+)\)`, `int($1.Weekday())`) // В Go воскресенье - это 0
	// Функция YEARDAY
	code = ReplaceAllStringRegexp(code, `yearday\(([^)]+)\)`, `$1.YearDay() - 1`) // В Go дни года начинаются с 1
	// Функция MAKETIME
	code = ReplaceAllStringRegexp(code, `maketime\(([^,]+),([^,]+),([^,]+),([^,]+),([^,]+),([^)]+)\)`,
		`time.Date($6 + 1, time.Month($5 + 1), $4, $1, $2, $3, 0, time.UTC).Unix()`)

	//функции времени выполнения
	// Функция CYCLESEC
	code = ReplaceAllStringRegexp(code, `(?i)\bcyclesec\(\)`, `CYCLESEC()`)
	// Функция EXECSEC
	code = ReplaceAllStringRegexp(code, `(?i)\bexecsec\(\)`, `EXECSEC()`)

	//функции над таймерами
	// Функция TIMERMSEC
	code = ReplaceAllStringRegexp(code, `timermsec\(([^)]+)\)`, `$1.Nanosecond() / 1e6`)
	// Функция TIMERSEC
	code = ReplaceAllStringRegexp(code, `timersec\(([^)]+)\)`, `$1.Second()`)
	// Функция TIMERMIN
	code = ReplaceAllStringRegexp(code, `timermin\(([^)]+)\)`, `$1.Minute()`)
	// Функция TIMERHOUR
	code = ReplaceAllStringRegexp(code, `timerhour\(([^)]+)\)`, `$1.Hour()`)
	// Функция MAKETIMER
	code = ReplaceAllStringRegexp(code, `maketimer\(([^,]+),([^,]+),([^,]+),([^)]+)\)`,
		`time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), $1, $2, $3, $4*1e6, time.Local)`)

	//функции счетчиков тиков
	//getticks
	code = ReplaceAllStringRegexp(code, `(?i)getticks`, "GETTICKS")
	//ticksize
	code = ReplaceAllStringRegexp(code, `(?i)ticksize`, "TICKSIZE")
	//set
	code = ReplaceAllStringRegexp(code, `SET\s+(\{[^{}]+\})\s*\[([^\]]+)\],\s*\(([^)]+)\)`, "PKG.UpdateVal($1[$2], $3, true)")

	code = ReplaceAllStringRegexp(code, `(?i)\bset\s+([^,]+(?:\{[^}]+\})?),\s*([^)\s]+)\s*(?:\)|\b)`, "PKG.UpdateVal($1, $2, true)\n")
	code = ReplaceAllStringRegexp(code, `(?i)\bset\s*{([^,]+(?:\{[^}]+\})?),\s*([^)\s]+)\s*(?:\)|\b)`, "PKG.UpdateVal({$1, $2, true)\n\t")

	//функции перезагрузки
	//stop_softdog
	code = ReplaceAllStringRegexp(code, `(?i)\bstop_softdog\(\)`, "STOP_SOFTDOG()")
	//reset
	code = ReplaceAllStringRegexp(code, `(?i)\breset\(([^)]+)\)`, "RESET($1)")

	//set_wait доделать!
	code = ReplaceAllStringRegexp(code, `(?i)set_wait`, "set_wait")
	//return
	code = ReplaceAllStringRegexp(code, `(?i)return`, "return")

	// FINDOUT с массивом aout
	code = ReplaceAllStringRegexp(code, `(?i)\bfindout\(\s*([^,]+),\s*([^,]+),\s*([^,]+)\s*\)`, "FINDOUT($1, $2, $3, aout)\n")

	//initouts
	code = ReplaceAllStringRegexp(code, `(?i)initouts\s+(\d+)\s*,\s*(\d+)\s*,\s*(\d+)\s*`, "INITOUTS($1, $2, $3)\n")
	//beep
	code = ReplaceAllStringRegexp(code, `(?i)beep\s*\(\s*\)\s*`, "BEEP()\n")
	//siren_on
	code = ReplaceAllStringRegexp(code, `(?i)siren_on\s*\(\s*\)\s*`, "SIREN_ON()\n")
	//siren_off
	code = ReplaceAllStringRegexp(code, `(?i)siren_off\s*\(\s*\)\s*`, "SIREN_OFF()\n")
	//execute
	code = ReplaceAllStringRegexp(code, `(?i)execute\s+\[?"?([^"\]\s]+)"?\]?\s*,\s*"([^"\s]+)"\s*`, "EXECUTE($1, \"$2\")\n")

	//sleep
	code = ReplaceAllStringRegexp(code, `sleep\(([^)]+)\)`, `time.Sleep(($1) * time.Second)`)

	//реперы
	Reps = findReps(code)
	code = ReplaceExpressions(code, Reps)
	code = ReplaceAllStringRegexp(code, `\.Value\[(.*?)\]`, `.$1`)

	code = strings.ReplaceAll(code, "x=0", "//x = 0")
	//условия
	//fmt.Println(code)
	code = ReplaceAllStringRegexp(code, `(?i)if\s*\((.+)\)`, `if $1 {`)
	//code = ReplaceAllStringRegexp(code, `(?i)\bIF\s*\(([^)]+)\)\s*(?![^{]*})`, `if ($1) {`)

	code = ReplaceAllStringRegexp(code, `(?i)ELSE`, "} else {")
	// Новая замена для cur0.Hour() на curhour()
	code = ReplaceAllStringRegexp(code, `cur0\.Hour\(\)`, `curhour()`)
	// Новая замена для [sys_num] на пустое значение
	code = ReplaceAllStringRegexp(code, `\[sys_num\]`, ``)

	//потому удалить, относится только к самой первой программе
	//code = strings.ReplaceAll(code, "dout[2]=2+(((Reps[\"ОХР КР ДЕС\"].Value) == (2)) && ((Reps[\"Вход ДЕС\"].Value) == (2)) && ((Reps[\"КРдоРУ ДЕС\"].Value) == (2)) && ((Reps[\"Выход ДЕС\"].Value) == (2)) && ((Reps[\"ВЫХ Д ДЕС\"].Value) == (2)))  // ход ао", "dout[2]=(2)+(((Reps[\"ОХР КР ДЕС\"].Value) == (2)) && ((Reps[\"Вход ДЕС\"].Value) == (2)) && ((Reps[\"КРдоРУ ДЕС\"].Value) == (2)) && ((Reps[\"Выход ДЕС\"].Value) == (2)) && ((Reps[\"ВЫХ Д ДЕС\"].Value) == (2)))  // ход ао\n    ")
	//code = strings.ReplaceAll(code, "dout[1]=2+(((Reps[\"ОХР КР ДЕС\"].Value) == (2)) && ((Reps[\"Вход ДЕС\"].Value) == (2)) && ((Reps[\"КРдоРУ ДЕС\"].Value) == (2)) && ((Reps[\"Выход ДЕС\"].Value) == (2)) && ((Reps[\"ВЫХ Д ДЕС\"].Value) == (2)) && ((Reps[\"СВзаВХ ДЕС\"].Value) == (1)) && ((Reps[\"СВдоВЫХ ДЕС\"].Value) == (1)) && ((Reps[\"СВ ОК ДЕС\"].Value) == (1)))  // ход ао", "dout[1]=(2)+(((Reps[\"ОХР КР ДЕС\"].Value) == (2)) && ((Reps[\"Вход ДЕС\"].Value) == (2)) && ((Reps[\"КРдоРУ ДЕС\"].Value) == (2)) && ((Reps[\"Выход ДЕС\"].Value) == (2)) && ((Reps[\"ВЫХ Д ДЕС\"].Value) == (2)) && ((Reps[\"СВзаВХ ДЕС\"].Value) == (1)) && ((Reps[\"СВдоВЫХ ДЕС\"].Value) == (1)) && ((Reps[\"СВ ОК ДЕС\"].Value) == (1)))  // ход ао")
	//code = strings.ReplaceAll(code, "(Reps[\"ЗадPгВыхРабДЕС\"].Value*1.15)", "((Reps[\"ЗадPгВыхРабДЕС\"].Value)*(1.15))")
	//code = strings.ReplaceAll(code, "return(SET_WAIT(sys,state,timeout))\n}", "return(SET_WAIT(sys,state,timeout))\n}\n return false")
	//code = strings.ReplaceAll(code, "  time.Sleep((5*18) * time.Second)\t// ждем первого опроса модулей", "  time.Sleep((5*18) * time.Second)\t// ждем первого опроса модулей\n return (t)")
	//для демонстрации
	code = strings.ReplaceAll(code, "//x = 0", "var x float32 \n\tx = 0")
	code = strings.ReplaceAll(code, "x=x || val(\"ПОЖАР ОПЕ КУШН\")", "x=x + val(\"ПОЖАР ОПЕ КУШН\")")
	code = strings.ReplaceAll(code, "x=x || val(\"ПОЖАР ПЕР КУШН\")", "x=x + val(\"ПОЖАР ПЕР КУШН\")")
	code = strings.ReplaceAll(code, "if ((valTrack(val(\"КН АВОСТ КУШН\")) == (4,8)),1) {", "if valTrack(val(\"КН АВОСТ КУШН\"), 4,8) == 1 {")
	code = strings.ReplaceAll(code, "aout[5]=TRUE(val(\"ДАТА АО КУШН\"))", "aout[5]=1")
	code = strings.ReplaceAll(code, "aout[6]=TRUE(val(\"ДАТА ЗАО КУШН\"))", "aout[6]=1")
	code = strings.ReplaceAll(code, "reason=checkPrecond(0)", "reason:=checkPrecond(0)")
	code = strings.ReplaceAll(code, "if reason) != (0 {", "if reason != 0 {")
	code = strings.ReplaceAll(code, "aout[5]=time.Now().Unix()", "aout[5]=float32(time.Now().Unix())")
	code = strings.ReplaceAll(code, "func oninit(t float32) float32 {", "func oninit() {")
	code = strings.ReplaceAll(code, "x=setwex(val", "setwex(")
	code = strings.ReplaceAll(code, "x=set_wait(val", "set_wait(")
	code = strings.ReplaceAll(code, "if val(\"КРАН ОХР КУШН\")) != (2 {", "if val(\"КРАН ОХР КУШН\") != 2 {")
	code = strings.ReplaceAll(code, "if checkFire(0) {", "if checkFire(0) > 0{")
	code = strings.ReplaceAll(code, "     //\n}", "")
	code = strings.ReplaceAll(code, "PKG.UpdateVal(val", "PKG.UpdateVal(")
	code = strings.ReplaceAll(code, "aout[6]=time.Now().Unix()", "aout[6]=float32(time.Now().Unix())")
	code = strings.ReplaceAll(code, "front((val(\"РЕЖИМ ГРС КУШН\")) != (0),9)", "front(\"РЕЖИМ ГРС КУШН\",9) > 0")
	//a6 pidreg
	code = strings.ReplaceAll(code, "x=x || ((val(\"КОНТР РЕГ КУШН\")) == (1))", "if ((val(\"КОНТР РЕГ КУШН\")) == (1)){ \n\t\tx += 1\n\t}")
	code = strings.ReplaceAll(code, "bpticks=0", "//bpticks=0")
	code = strings.ReplaceAll(code, "prevhour=curhour()", "//prevhour=curhour()")
	code = strings.ReplaceAll(code, "mode=checkPrecond(0)", "mode:=checkPrecond(0)")
	code = strings.ReplaceAll(code, "x=setSens(val", "setSens(")
	code = strings.ReplaceAll(code, "u=pid(1,val(\"ПИДСАУ КИ КУШН\"),val(\"ПИДСАУ КП КУШН\"),val(\"РВЫХ ЗАД КУШН\"),val(\"ПОЗ ЗАДВ КУШН\"),val(\"РВЫХ 123 КУШН\"),DOST(РВЫХ 123 КУШН),mode)", "var k float32\nk = 0\nif dost(\"РВЫХ 123 КУШН\") {\n\tk = 1\n}\nu:=pid(1,val(\"ПИДСАУ КИ КУШН\"),val(\"ПИДСАУ КП КУШН\"),val(\"РВЫХ ЗАД КУШН\"),val(\"ПОЗ ЗАДВ КУШН\"),val(\"РВЫХ 123 КУШН\"),k,mode)")

	return code
}

func main() {
	code := `package main

import (
	PKG "AlgorithmsRabbit/connections"
	"sync"
	"time"
)

var aout [100]float32
var dout [100]float32
	// Время запуска программы
var startTime = time.Now()
type SafeMap struct {
	Mu   sync.Mutex
	Reps map[string]*Rep
}

type Rep struct {
	MEK_Address int
	Raper       string
	Value       float32
	TypeParam   string
	OldValue    float32
	Reliability bool
	TimeOld     time.Time
	Time        time.Time
}

type OutToRabbitMQ struct {
	MEK_Address int
	Raper       string
	Value       float32
	TypeParam   string
	Reliability bool
	Time        time.Time
}

func main() {
	PKG.CONNECTRABBITMIB = "amqp://admin:admin@127.0.0.1:5672/"
	PKG.NameAlg = "ButtonALG"
	//Объявление входных и выходных массивов
	PKG.DeclareArrays()
	//Подключаемся к RabbitMQ
	PKG.DeclareRabbit()
	//Запрашиваем и отправляем данные
	go PKG.ConsumeFromRabbitMq(&PKG.InputMap)
	go PKG.SendToRabbitMQ(&PKG.OutputMap)
	for {
		//Если данные получены, начинаем алгоритм
		if PKG.ConnectToRabit {
			for {
				mainOutput()
				time.Sleep(200 * time.Millisecond)
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
}
	`

	// Чтение данных из файла
	fileContent, err := ioutil.ReadFile("input.txt")
	if err != nil {
		fmt.Println("Ошибка чтения файла:", err)
		return
	}

	// Сохранение данных в переменной code
	inputCode := string(fileContent)
	outputCode1 := translate_for_to_go(inputCode)
	outputCode := addBracesToIfStatements(outputCode1)
	codeFinal := code + outputCode + "\n}"

	// Запись данных в файл "output.go"
	err = ioutil.WriteFile("output.go", []byte(codeFinal), 0644)
	if err != nil {
		fmt.Println("Ошибка записи в файл:", err)
		return
	}

	// Запись списка всех реперов (мап Reps) в файл "reps.txt"
	repsContent := ""
	for key := range Reps {
		repsContent += fmt.Sprintf("[[Algorithm.Params]]\nRaper = \"%s\"\nTypeParam = \"IN\"\n", key)
	}
	err = ioutil.WriteFile("reps.txt", []byte(repsContent), 0644)
	if err != nil {
		fmt.Println("Ошибка записи в файл reps.txt:", err)
		return
	}

	fmt.Println("Операции чтения из файла и записи в файл успешно выполнены.")

}
