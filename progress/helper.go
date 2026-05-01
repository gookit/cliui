package progress

import (
	"math/rand/v2"
)

func repeatRune(char rune, length int) (chars []rune) {
	for range length {
		chars = append(chars, char)
	}
	return
}

// CharThemes collection. can use for Progress bar, RoundTripSpinner
var CharThemes = []rune{
	CharEqual,
	CharCenter,
	CharSquare,
	CharSquare1,
	CharSquare2,
}

// GetCharTheme by index number. if index not exist, will return a random theme
func GetCharTheme(index int) rune {
	if index > 0 && len(CharThemes) > index {
		return CharThemes[index]
	}
	return RandomCharTheme()
}

// RandomCharTheme get
func RandomCharTheme() rune {
	return CharThemes[rand.IntN(len(CharThemes)-1)]
}

// CharsThemes collection. can use for LoadingBar, LoadingSpinner
var CharsThemes = [][]rune{
	{'卍', '卐'},
	{'☺', '☻'},
	{'░', '▒', '▓'},
	{'⊘', '⊖', '⊕', '⊗'},
	{'◐', '◒', '◓', '◑'},
	{'✣', '✤', '✥', '❉'},
	{'-', '\\', '|', '/'},
	{'▢', '■', '▢', '■'},
	[]rune("▖▘▝▗"),
	[]rune("◢◣◤◥"),
	[]rune("⌞⌟⌝⌜"),
	[]rune("◎●◯◌○⊙"),
	[]rune("◡◡⊙⊙◠◠"),
	[]rune("⇦⇧⇨⇩"),
	[]rune("✳✴✵✶✷✸✹"),
	[]rune("←↖↑↗→↘↓↙"),
	[]rune("➩➪➫➬➭➮➯➱"),
	[]rune("①②③④"),
	[]rune("㊎㊍㊌㊋㊏"),
	[]rune("⣾⣽⣻⢿⡿⣟⣯⣷"),
	[]rune("⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏"),
	[]rune("▉▊▋▌▍▎▏▎▍▌▋▊▉"),
	[]rune("🌍🌎🌏"),
	[]rune("☰☱☲☳☴☵☶☷"),
	[]rune("⠋⠙⠚⠒⠂⠂⠒⠲⠴⠦⠖⠒⠐⠐⠒⠓⠋"),
	[]rune("🕐🕑🕒🕓🕔🕕🕖🕗🕘🕙🕚🕛"),
}

// GetCharsTheme by index number
func GetCharsTheme(index int) []rune {
	if index > 0 && len(CharsThemes) > index {
		return CharsThemes[index]
	}
	return RandomCharsTheme()
}

// RandomCharsTheme get
func RandomCharsTheme() []rune {
	return CharsThemes[rand.IntN(len(CharsThemes)-1)]
}

func normalizeMaxSteps(maxSteps int64) int64 {
	if maxSteps < 0 {
		return 0
	}
	return maxSteps
}
