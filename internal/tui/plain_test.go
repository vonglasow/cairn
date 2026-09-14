package tui

// plainSkin is a skin with no colour in it, so a test can compare what a card
// says rather than what it looks like. Styling is checked where it is decided —
// in the pinned frames — and nowhere else.
var plainSkin = Skin{palette: Default}
