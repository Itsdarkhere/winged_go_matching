package registration

/*
	Test helpers (only used in tests)
*/

// SetTexter sets the texter dependency for the Business instance.
func (b *Business) SetTexter(t texter) {
	b.texter = t
}

// SetStorer sets the storer dependency for the Business instance.
func (b *Business) SetStorer(s storer) {
	b.storer = s
}

// SetBEUploader sets the beUploader dependency for the Business instance.
func (b *Business) SetBEUploader(u uploader) {
	b.beUploader = u
}

// SetAIUploader sets the aiUploader dependency for the Business instance.
func (b *Business) SetAIUploader(u uploader) {
	b.aiUploader = u
}
