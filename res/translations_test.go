package res

import "testing"

func TestTranslationFilesExist(t *testing.T) {
	translationFiles := []string{
		"translations/zh-Hans.json",
		"translations/zh-Hant.json",
		"translations/en.json",
	}

	for _, fileName := range translationFiles {
		t.Run(fileName, func(t *testing.T) {
			if _, err := Translations.ReadFile(fileName); err != nil {
				t.Fatalf("Translations.ReadFile(%q) returned error: %v", fileName, err)
			}
		})
	}
}

func TestTranslationsInfoRegistration(t *testing.T) {
	tests := []struct {
		name                string
		translationFileName string
	}{
		{name: "zhHans", translationFileName: "zh-Hans.json"},
		{name: "zhHant", translationFileName: "zh-Hant.json"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			found := false
			for _, info := range TranslationsInfo {
				if info.Name == tt.name && info.TranslationFileName == tt.translationFileName {
					found = true
					break
				}
			}

			if !found {
				t.Fatalf("TranslationsInfo missing entry with Name=%q and TranslationFileName=%q", tt.name, tt.translationFileName)
			}
		})
	}
}
