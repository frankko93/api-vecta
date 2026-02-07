package config

// UnitOfMeasure represents units for mineral measurements
type UnitOfMeasure string

const (
	UnitTonnes     UnitOfMeasure = "tonnes"
	UnitKilograms  UnitOfMeasure = "kilograms"
	UnitGrams      UnitOfMeasure = "grams"
	UnitTroyOunces UnitOfMeasure = "troy_ounces"
)

// IsValid validates if the unit is supported
func (u UnitOfMeasure) IsValid() bool {
	switch u {
	case UnitTonnes, UnitKilograms, UnitGrams, UnitTroyOunces:
		return true
	}
	return false
}

// GetAvailableUnits returns all available units
func GetAvailableUnits() []map[string]string {
	return []map[string]string{
		{"value": string(UnitTonnes), "label": "Toneladas"},
		{"value": string(UnitKilograms), "label": "Kilogramos"},
		{"value": string(UnitGrams), "label": "Gramos"},
		{"value": string(UnitTroyOunces), "label": "Onzas Troy"},
	}
}
