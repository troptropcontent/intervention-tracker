package translation

type TranslatorService interface {
	MustTranslate(key string, templateData ...map[string]interface{}) string
}
