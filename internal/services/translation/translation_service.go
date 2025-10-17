package translation

import (
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/pelletier/go-toml/v2"
	"github.com/troptropcontent/qr_code_maintenance/internal/utils"
	"golang.org/x/text/language"
)

type Translator struct {
	localizer *i18n.Localizer
}

func NewTranslator() *Translator {
	bundle := i18n.NewBundle(language.French)
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
	localFilePath := utils.MustGetPathFromRoot("internal/locales/fr.toml")
	bundle.MustLoadMessageFile(localFilePath)
	localizer := i18n.NewLocalizer(bundle, language.French.String())

	return &Translator{
		localizer: localizer,
	}
}

func (t *Translator) MustTranslate(key string, templateData ...map[string]interface{}) string {
	data := make(map[string]interface{})
	if len(templateData) > 0 {
		data = templateData[0]
	}
	translation := t.localizer.MustLocalize(&i18n.LocalizeConfig{
		MessageID:    key,
		TemplateData: data,
		PluralCount:  data["Count"],
	})
	return translation
}
