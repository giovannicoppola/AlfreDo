package parser

import (
	"strings"
	"time"
)

// localeKeyword maps a keyword to either a relative day offset or a weekday.
type localeKeyword struct {
	days    int            // relative days from today (0=today, 1=tomorrow, etc.)
	weekday time.Weekday   // target weekday (only used if isWeekday is true)
	isWeekday bool
}

// localeKeywords maps language code → lowercase keyword → resolution.
// Covers Todoist-supported languages: da, de, en, es, fi, fr, it, ja, ko, nl, pl, pt, ru, sv, tr, zh.
var localeKeywords = map[string]map[string]localeKeyword{
	"it": {
		// relative
		"oggi":        {days: 0},
		"domani":      {days: 1},
		"dopodomani":  {days: 2},
		// weekdays
		"lunedì":    {weekday: time.Monday, isWeekday: true},
		"lunedi":    {weekday: time.Monday, isWeekday: true},
		"martedì":   {weekday: time.Tuesday, isWeekday: true},
		"martedi":   {weekday: time.Tuesday, isWeekday: true},
		"mercoledì": {weekday: time.Wednesday, isWeekday: true},
		"mercoledi": {weekday: time.Wednesday, isWeekday: true},
		"giovedì":   {weekday: time.Thursday, isWeekday: true},
		"giovedi":   {weekday: time.Thursday, isWeekday: true},
		"venerdì":   {weekday: time.Friday, isWeekday: true},
		"venerdi":   {weekday: time.Friday, isWeekday: true},
		"sabato":    {weekday: time.Saturday, isWeekday: true},
		"domenica":  {weekday: time.Sunday, isWeekday: true},
	},
	"de": {
		"heute":       {days: 0},
		"morgen":      {days: 1},
		"übermorgen":  {days: 2},
		"ubermorgen":  {days: 2},
		"montag":      {weekday: time.Monday, isWeekday: true},
		"dienstag":    {weekday: time.Tuesday, isWeekday: true},
		"mittwoch":    {weekday: time.Wednesday, isWeekday: true},
		"donnerstag":  {weekday: time.Thursday, isWeekday: true},
		"freitag":     {weekday: time.Friday, isWeekday: true},
		"samstag":     {weekday: time.Saturday, isWeekday: true},
		"sonntag":     {weekday: time.Sunday, isWeekday: true},
	},
	"fr": {
		"aujourd'hui": {days: 0},
		"aujourdhui":  {days: 0},
		"demain":      {days: 1},
		"après-demain": {days: 2},
		"apres-demain": {days: 2},
		"lundi":       {weekday: time.Monday, isWeekday: true},
		"mardi":       {weekday: time.Tuesday, isWeekday: true},
		"mercredi":    {weekday: time.Wednesday, isWeekday: true},
		"jeudi":       {weekday: time.Thursday, isWeekday: true},
		"vendredi":    {weekday: time.Friday, isWeekday: true},
		"samedi":      {weekday: time.Saturday, isWeekday: true},
		"dimanche":    {weekday: time.Sunday, isWeekday: true},
	},
	"es": {
		"hoy":           {days: 0},
		"mañana":        {days: 1},
		"manana":        {days: 1},
		"pasado mañana": {days: 2},
		"pasado manana": {days: 2},
		"lunes":         {weekday: time.Monday, isWeekday: true},
		"martes":        {weekday: time.Tuesday, isWeekday: true},
		"miércoles":     {weekday: time.Wednesday, isWeekday: true},
		"miercoles":     {weekday: time.Wednesday, isWeekday: true},
		"jueves":        {weekday: time.Thursday, isWeekday: true},
		"viernes":       {weekday: time.Friday, isWeekday: true},
		"sábado":        {weekday: time.Saturday, isWeekday: true},
		"sabado":        {weekday: time.Saturday, isWeekday: true},
		"domingo":       {weekday: time.Sunday, isWeekday: true},
	},
	"pt": {
		"hoje":          {days: 0},
		"amanhã":        {days: 1},
		"amanha":        {days: 1},
		"depois de amanhã": {days: 2},
		"depois de amanha": {days: 2},
		"segunda-feira": {weekday: time.Monday, isWeekday: true},
		"segunda":       {weekday: time.Monday, isWeekday: true},
		"terça-feira":   {weekday: time.Tuesday, isWeekday: true},
		"terca-feira":   {weekday: time.Tuesday, isWeekday: true},
		"terça":         {weekday: time.Tuesday, isWeekday: true},
		"terca":         {weekday: time.Tuesday, isWeekday: true},
		"quarta-feira":  {weekday: time.Wednesday, isWeekday: true},
		"quarta":        {weekday: time.Wednesday, isWeekday: true},
		"quinta-feira":  {weekday: time.Thursday, isWeekday: true},
		"quinta":        {weekday: time.Thursday, isWeekday: true},
		"sexta-feira":   {weekday: time.Friday, isWeekday: true},
		"sexta":         {weekday: time.Friday, isWeekday: true},
		"sábado":        {weekday: time.Saturday, isWeekday: true},
		"sabado":        {weekday: time.Saturday, isWeekday: true},
		"domingo":       {weekday: time.Sunday, isWeekday: true},
	},
	"nl": {
		"vandaag":      {days: 0},
		"morgen":       {days: 1},
		"overmorgen":   {days: 2},
		"maandag":      {weekday: time.Monday, isWeekday: true},
		"dinsdag":      {weekday: time.Tuesday, isWeekday: true},
		"woensdag":     {weekday: time.Wednesday, isWeekday: true},
		"donderdag":    {weekday: time.Thursday, isWeekday: true},
		"vrijdag":      {weekday: time.Friday, isWeekday: true},
		"zaterdag":     {weekday: time.Saturday, isWeekday: true},
		"zondag":       {weekday: time.Sunday, isWeekday: true},
	},
	"da": {
		"i dag":        {days: 0},
		"idag":         {days: 0},
		"i morgen":     {days: 1},
		"imorgen":      {days: 1},
		"i overmorgen": {days: 2},
		"mandag":       {weekday: time.Monday, isWeekday: true},
		"tirsdag":      {weekday: time.Tuesday, isWeekday: true},
		"onsdag":       {weekday: time.Wednesday, isWeekday: true},
		"torsdag":      {weekday: time.Thursday, isWeekday: true},
		"fredag":       {weekday: time.Friday, isWeekday: true},
		"lørdag":       {weekday: time.Saturday, isWeekday: true},
		"lordag":       {weekday: time.Saturday, isWeekday: true},
		"søndag":       {weekday: time.Sunday, isWeekday: true},
		"sondag":       {weekday: time.Sunday, isWeekday: true},
	},
	"sv": {
		"idag":         {days: 0},
		"imorgon":      {days: 1},
		"i morgon":     {days: 1},
		"övermorgon":   {days: 2},
		"overmorgon":   {days: 2},
		"måndag":       {weekday: time.Monday, isWeekday: true},
		"mandag":       {weekday: time.Monday, isWeekday: true},
		"tisdag":       {weekday: time.Tuesday, isWeekday: true},
		"onsdag":       {weekday: time.Wednesday, isWeekday: true},
		"torsdag":      {weekday: time.Thursday, isWeekday: true},
		"fredag":       {weekday: time.Friday, isWeekday: true},
		"lördag":       {weekday: time.Saturday, isWeekday: true},
		"lordag":       {weekday: time.Saturday, isWeekday: true},
		"söndag":       {weekday: time.Sunday, isWeekday: true},
		"sondag":       {weekday: time.Sunday, isWeekday: true},
	},
	"fi": {
		"tänään":       {days: 0},
		"tanaan":       {days: 0},
		"huomenna":     {days: 1},
		"ylihuomenna":  {days: 2},
		"maanantai":    {weekday: time.Monday, isWeekday: true},
		"tiistai":      {weekday: time.Tuesday, isWeekday: true},
		"keskiviikko":  {weekday: time.Wednesday, isWeekday: true},
		"torstai":      {weekday: time.Thursday, isWeekday: true},
		"perjantai":    {weekday: time.Friday, isWeekday: true},
		"lauantai":     {weekday: time.Saturday, isWeekday: true},
		"sunnuntai":    {weekday: time.Sunday, isWeekday: true},
	},
	"pl": {
		"dzisiaj":      {days: 0},
		"dziś":         {days: 0},
		"dzis":         {days: 0},
		"jutro":        {days: 1},
		"pojutrze":     {days: 2},
		"poniedziałek": {weekday: time.Monday, isWeekday: true},
		"poniedzialek": {weekday: time.Monday, isWeekday: true},
		"wtorek":       {weekday: time.Tuesday, isWeekday: true},
		"środa":        {weekday: time.Wednesday, isWeekday: true},
		"sroda":        {weekday: time.Wednesday, isWeekday: true},
		"czwartek":     {weekday: time.Thursday, isWeekday: true},
		"piątek":       {weekday: time.Friday, isWeekday: true},
		"piatek":       {weekday: time.Friday, isWeekday: true},
		"sobota":       {weekday: time.Saturday, isWeekday: true},
		"niedziela":    {weekday: time.Sunday, isWeekday: true},
	},
	"ru": {
		"сегодня":      {days: 0},
		"завтра":       {days: 1},
		"послезавтра":  {days: 2},
		"понедельник":  {weekday: time.Monday, isWeekday: true},
		"вторник":      {weekday: time.Tuesday, isWeekday: true},
		"среда":        {weekday: time.Wednesday, isWeekday: true},
		"четверг":      {weekday: time.Thursday, isWeekday: true},
		"пятница":      {weekday: time.Friday, isWeekday: true},
		"суббота":      {weekday: time.Saturday, isWeekday: true},
		"воскресенье":  {weekday: time.Sunday, isWeekday: true},
	},
	"tr": {
		"bugün":        {days: 0},
		"bugun":        {days: 0},
		"yarın":        {days: 1},
		"yarin":        {days: 1},
		"öbür gün":     {days: 2},
		"obur gun":     {days: 2},
		"pazartesi":    {weekday: time.Monday, isWeekday: true},
		"salı":         {weekday: time.Tuesday, isWeekday: true},
		"sali":         {weekday: time.Tuesday, isWeekday: true},
		"çarşamba":     {weekday: time.Wednesday, isWeekday: true},
		"carsamba":     {weekday: time.Wednesday, isWeekday: true},
		"perşembe":     {weekday: time.Thursday, isWeekday: true},
		"persembe":     {weekday: time.Thursday, isWeekday: true},
		"cuma":         {weekday: time.Friday, isWeekday: true},
		"cumartesi":    {weekday: time.Saturday, isWeekday: true},
		"pazar":        {weekday: time.Sunday, isWeekday: true},
	},
	"ja": {
		"今日":   {days: 0},
		"きょう": {days: 0},
		"明日":   {days: 1},
		"あした": {days: 1},
		"明後日": {days: 2},
		"あさって": {days: 2},
		"月曜日": {weekday: time.Monday, isWeekday: true},
		"火曜日": {weekday: time.Tuesday, isWeekday: true},
		"水曜日": {weekday: time.Wednesday, isWeekday: true},
		"木曜日": {weekday: time.Thursday, isWeekday: true},
		"金曜日": {weekday: time.Friday, isWeekday: true},
		"土曜日": {weekday: time.Saturday, isWeekday: true},
		"日曜日": {weekday: time.Sunday, isWeekday: true},
	},
	"ko": {
		"오늘":   {days: 0},
		"내일":   {days: 1},
		"모레":   {days: 2},
		"월요일": {weekday: time.Monday, isWeekday: true},
		"화요일": {weekday: time.Tuesday, isWeekday: true},
		"수요일": {weekday: time.Wednesday, isWeekday: true},
		"목요일": {weekday: time.Thursday, isWeekday: true},
		"금요일": {weekday: time.Friday, isWeekday: true},
		"토요일": {weekday: time.Saturday, isWeekday: true},
		"일요일": {weekday: time.Sunday, isWeekday: true},
	},
	"zh": {
		"今天": {days: 0},
		"明天": {days: 1},
		"后天": {days: 2},
		"後天": {days: 2},
		"星期一": {weekday: time.Monday, isWeekday: true},
		"星期二": {weekday: time.Tuesday, isWeekday: true},
		"星期三": {weekday: time.Wednesday, isWeekday: true},
		"星期四": {weekday: time.Thursday, isWeekday: true},
		"星期五": {weekday: time.Friday, isWeekday: true},
		"星期六": {weekday: time.Saturday, isWeekday: true},
		"星期日": {weekday: time.Sunday, isWeekday: true},
		"周一":   {weekday: time.Monday, isWeekday: true},
		"周二":   {weekday: time.Tuesday, isWeekday: true},
		"周三":   {weekday: time.Wednesday, isWeekday: true},
		"周四":   {weekday: time.Thursday, isWeekday: true},
		"周五":   {weekday: time.Friday, isWeekday: true},
		"周六":   {weekday: time.Saturday, isWeekday: true},
		"周日":   {weekday: time.Sunday, isWeekday: true},
	},
}

// localeNextPrefixes maps language code → list of "next" prefixes to strip.
// e.g., "nächsten Freitag" → strip "nächsten" → look up "freitag".
var localeNextPrefixes = map[string][]string{
	"it": {"prossimo", "prossima"},
	"de": {"nächsten", "nächste", "nächstem", "nachsten", "nachste", "nachstem"},
	"fr": {"prochain", "prochaine"},
	"es": {"próximo", "próxima", "proximo", "proxima"},
	"pt": {"próximo", "próxima", "proximo", "proxima"},
	"nl": {"volgende", "aanstaande"},
	"da": {"næste", "naeste"},
	"sv": {"nästa", "nasta"},
	"fi": {"ensi", "seuraava", "seuraavan"},
	"pl": {"następny", "następna", "następne", "nastepny", "nastepna", "nastepne"},
	"ru": {"следующий", "следующая", "следующее"},
	"tr": {"gelecek", "önümüzdeki", "onumuzdeki"},
	"ja": {"次の", "来週の"},
	"ko": {"다음"},
	"zh": {"下个", "下個"},
}

// ParseLocaleDateKeyword tries to resolve a date keyword using locale-specific tables.
// lang is a 2-letter language code (e.g., "it", "de").
// Returns the resolved time and true if the keyword was recognized.
func ParseLocaleDateKeyword(input, lang string) (time.Time, bool) {
	if lang == "" || lang == "en" {
		return time.Time{}, false // English handled by olebedev/when
	}

	keywords, ok := localeKeywords[lang]
	if !ok {
		return time.Time{}, false
	}

	normalized := strings.ToLower(strings.TrimSpace(input))

	// Direct lookup
	kw, found := keywords[normalized]
	if !found {
		// Try stripping "next" prefixes (e.g., "nächsten freitag" → "freitag")
		if prefixes, ok := localeNextPrefixes[lang]; ok {
			for _, prefix := range prefixes {
				after, stripped := strings.CutPrefix(normalized, strings.ToLower(prefix)+" ")
				if stripped {
					kw, found = keywords[strings.TrimSpace(after)]
					if found {
						break
					}
				}
			}
		}
	}

	if !found {
		return time.Time{}, false
	}

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	if kw.isWeekday {
		return nextWeekday(today, kw.weekday), true
	}
	return today.AddDate(0, 0, kw.days), true
}

// nextWeekday returns the next occurrence of the given weekday.
// If today is that weekday, returns 7 days from now (next week).
func nextWeekday(from time.Time, target time.Weekday) time.Time {
	current := from.Weekday()
	daysAhead := int(target) - int(current)
	if daysAhead <= 0 {
		daysAhead += 7
	}
	return from.AddDate(0, 0, daysAhead)
}
