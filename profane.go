package main

import "strings"

func profaneCheck(body string, profaneWords []string) string {
	words := strings.Split(body, " ")

	for i, word := range words {
		for _, profaneWord := range profaneWords {
			if strings.ToLower(word) == profaneWord {
				words[i] = "****"
			}
		}
	}
	return strings.Join(words, " ")
}
