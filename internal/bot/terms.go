// File: internal/bot/terms.go

package bot

import (
	"fmt"
	"log"
	"strings"

	"gorm.io/gorm"
)

// AddTerm adds a new term and its description to the database.
func (b *Bot) AddTerm(term, description string) error {
	// No need to replace if newlines are already there; store the markdown as-is
	newTerm := Term{
		Term:        term,
		Description: description,
	}
	result := b.DB.Create(&newTerm)
	if result.Error != nil {
		log.Printf("Error adding term '%s': %v", term, result.Error)
		return result.Error
	}
	log.Printf("Successfully added term '%s'", term)
	return nil
}

// EditTerm updates an existing term with a new description.
func (b *Bot) EditTerm(term, newDescription string) error {
	// Replace literal `\n` with actual newline characters to store properly formatted text
	newDescription = strings.ReplaceAll(newDescription, `\n`, "\n")

	// Update the term in the database with the formatted description
	result := b.DB.Model(&Term{}).Where("term = ?", term).Update("description", newDescription)
	if result.Error != nil {
		log.Printf("Error editing term '%s': %v", term, result.Error)
		return result.Error
	}
	log.Printf("Successfully updated term '%s'", term)
	return nil
}

// RemoveTerm removes a term from the database.
func (b *Bot) RemoveTerm(term string) error {
	result := b.DB.Where("term = ?", term).Delete(&Term{})
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			log.Printf("Term '%s' not found for removal", term)
			return fmt.Errorf("term '%s' not found", term)
		}
		log.Printf("Error removing term '%s': %v", term, result.Error)
		return result.Error
	}

	if result.RowsAffected == 0 {
		log.Printf("No term found with name '%s' to remove", term)
		return fmt.Errorf("term '%s' not found", term)
	}

	log.Printf("Successfully removed term '%s'", term)
	return nil
}

// ListTerms lists all terms in the database.
func (b *Bot) ListTerms() ([]Term, error) {
	var terms []Term
	result := b.DB.Order("term").Find(&terms)
	if result.Error != nil {
		log.Printf("Error listing terms: %v", result.Error)
		return nil, result.Error
	}
	log.Printf("Successfully listed %d terms", len(terms))
	return terms, nil
}

// GetTermDescription gets the description of a given term.
func (b *Bot) GetTermDescription(term string) (string, error) {
	var t Term
	result := b.DB.Where("term = ?", term).First(&t)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			log.Printf("Term '%s' not found when getting description", term)
			return "", fmt.Errorf("term '%s' not found", term)
		}
		log.Printf("Error retrieving term '%s' description: %v", term, result.Error)
		return "", result.Error
	}
	log.Printf("Successfully retrieved description for term '%s'", term)
	return t.Description, nil
}
