package bot

import (
	"regexp"
	"testing"

	"github.com/go-telegram/bot/models"
)

func findHandler(b *Bot, id string) *handler {
	b.handlersMx.RLock()
	defer b.handlersMx.RUnlock()

	for _, h := range b.handlers {
		if h.id == id {
			return &h
		}
	}

	return nil
}

func Test_match_func(t *testing.T) {
	b := &Bot{}

	var called bool

	id := b.RegisterHandlerMatchFunc(func(update *models.Update) bool {
		called = true
		if update.ID != 42 {
			t.Error("invalid update id")
		}
		return true
	}, nil)

	h := findHandler(b, id)

	res := h.match(&models.Update{ID: 42})
	if !called {
		t.Error("not called")
	}
	if !res {
		t.Error("unexpected false result")
	}
}

func Test_match_exact(t *testing.T) {
	b := &Bot{}

	// Use regex for exact match
	id := b.RegisterHandler(HandlerTypeMessageText, "^xxx$", nil)

	h := findHandler(b, id)

	res := h.match(&models.Update{Message: &models.Message{Text: "zzz"}})
	if res {
		t.Error("unexpected true result")
	}

	res = h.match(&models.Update{Message: &models.Message{Text: "xxx"}})
	if !res {
		t.Error("unexpected false result")
	}
}

func Test_match_caption_exact(t *testing.T) {
	b := &Bot{}

	// Use regex for exact match
	id := b.RegisterHandler(HandlerTypePhotoCaption, "^xxx$", nil)

	h := findHandler(b, id)

	res := h.match(&models.Update{Message: &models.Message{Caption: "zzz"}})
	if res {
		t.Error("unexpected true result")
	}

	res = h.match(&models.Update{Message: &models.Message{Caption: "xxx"}})
	if !res {
		t.Error("unexpected false result")
	}
}

func Test_match_prefix(t *testing.T) {
	b := &Bot{}

	// Use regex for prefix match
	id := b.RegisterHandler(HandlerTypeCallbackQueryData, "^abc", nil)

	h := findHandler(b, id)

	res := h.match(&models.Update{CallbackQuery: &models.CallbackQuery{Data: "xabcdef"}})
	if res {
		t.Error("unexpected true result")
	}

	res = h.match(&models.Update{CallbackQuery: &models.CallbackQuery{Data: "abcdef"}})
	if !res {
		t.Error("unexpected false result")
	}
}

func Test_match_contains(t *testing.T) {
	b := &Bot{}

	// String pattern uses contains matching
	id := b.RegisterHandler(HandlerTypeCallbackQueryData, "abc", nil)

	h := findHandler(b, id)

	res := h.match(&models.Update{CallbackQuery: &models.CallbackQuery{Data: "xxabxx"}})
	if res {
		t.Error("unexpected true result")
	}

	res = h.match(&models.Update{CallbackQuery: &models.CallbackQuery{Data: "xxabcdef"}})
	if !res {
		t.Error("unexpected false result")
	}
}

func Test_match_regexp(t *testing.T) {
	b := &Bot{}

	re := regexp.MustCompile("^[a-z]+$")

	// Pass *regexp.Regexp directly as pattern
	id := b.RegisterHandler(HandlerTypeCallbackQueryData, re, nil)

	h := findHandler(b, id)

	res := h.match(&models.Update{CallbackQuery: &models.CallbackQuery{Data: "123abc"}})
	if res {
		t.Error("unexpected true result")
	}

	res = h.match(&models.Update{CallbackQuery: &models.CallbackQuery{Data: "abcdef"}})
	if !res {
		t.Error("unexpected false result")
	}
}

func Test_match_invalid_type(t *testing.T) {
	b := &Bot{}

	id := b.RegisterHandler(-1, "", nil)

	h := findHandler(b, id)

	res := h.match(&models.Update{CallbackQuery: &models.CallbackQuery{Data: "123abc"}})
	if res {
		t.Error("unexpected true result")
	}
}

func TestBot_RegisterUnregisterHandler(t *testing.T) {
	b := &Bot{}

	id1 := b.RegisterHandler(HandlerTypeCallbackQueryData, "", nil)
	id2 := b.RegisterHandler(HandlerTypeCallbackQueryData, "", nil)

	if len(b.handlers) != 2 {
		t.Fatalf("unexpected handlers len")
	}
	if h := findHandler(b, id1); h == nil {
		t.Fatalf("handler not found")
	}
	if h := findHandler(b, id2); h == nil {
		t.Fatalf("handler not found")
	}

	b.UnregisterHandler(id1)
	if len(b.handlers) != 1 {
		t.Fatalf("unexpected handlers len")
	}
	if h := findHandler(b, id1); h != nil {
		t.Fatalf("handler found")
	}
	if h := findHandler(b, id2); h == nil {
		t.Fatalf("handler not found")
	}
}

func Test_match_exact_game(t *testing.T) {
	b := &Bot{}

	// Use regex for exact match
	id := b.RegisterHandler(HandlerTypeCallbackQueryGameShortName, "^xxx$", nil)

	h := findHandler(b, id)
	u := models.Update{
		ID: 42,
		CallbackQuery: &models.CallbackQuery{
			ID:            "1000",
			GameShortName: "xxx",
		},
	}

	res := h.match(&u)
	if !res {
		t.Error("unexpected false result")
	}
}

func Test_match_command(t *testing.T) {
	t.Run("command at start, yes", func(t *testing.T) {
		b := &Bot{}

		// Use HandlerTypeCommand for command matching
		id := b.RegisterHandler(HandlerTypeCommand, "foo", nil)

		h := findHandler(b, id)
		u := models.Update{
			ID: 42,
			Message: &models.Message{
				Text: "/foo",
				Entities: []models.MessageEntity{
					{Type: models.MessageEntityTypeBotCommand, Offset: 0, Length: 4},
				},
			},
		}

		res := h.match(&u)
		if !res {
			t.Error("unexpected result")
		}
	})

	t.Run("command not at start, no", func(t *testing.T) {
		b := &Bot{}

		id := b.RegisterHandler(HandlerTypeCommand, "foo", nil)

		h := findHandler(b, id)
		u := models.Update{
			ID: 42,
			Message: &models.Message{
				Text: "a /foo",
				Entities: []models.MessageEntity{
					{Type: models.MessageEntityTypeBotCommand, Offset: 2, Length: 4},
				},
			},
		}

		res := h.match(&u)
		if res {
			t.Error("unexpected result - commands should only match at start")
		}
	})

	t.Run("wrong command, no", func(t *testing.T) {
		b := &Bot{}

		id := b.RegisterHandler(HandlerTypeCommand, "foo", nil)

		h := findHandler(b, id)
		u := models.Update{
			ID: 42,
			Message: &models.Message{
				Text: "/bar",
				Entities: []models.MessageEntity{
					{Type: models.MessageEntityTypeBotCommand, Offset: 0, Length: 4},
				},
			},
		}

		res := h.match(&u)
		if res {
			t.Error("unexpected result")
		}
	})

	t.Run("command at start with args, yes", func(t *testing.T) {
		b := &Bot{}

		id := b.RegisterHandler(HandlerTypeCommand, "foo", nil)

		h := findHandler(b, id)
		u := models.Update{
			ID: 42,
			Message: &models.Message{
				Text: "/foo arg1 arg2",
				Entities: []models.MessageEntity{
					{Type: models.MessageEntityTypeBotCommand, Offset: 0, Length: 4},
				},
			},
		}

		res := h.match(&u)
		if !res {
			t.Error("unexpected result")
		}
	})
}
