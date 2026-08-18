package services

import (
	"errors"
	"testing"

	clientsapi "github.com/jimjibone/woodhouse-api/go/v1/clients"
)

func nameAttrPerms(t *testing.T, srv *Info) clientsapi.Permissions {
	t.Helper()
	for _, attr := range srv.Pb().GetAttrs() {
		if attr.GetId() == "name" {
			return attr.GetText().GetPerms()
		}
	}
	t.Fatal("name attribute not found in service pb")
	return clientsapi.Permissions_PERM_UNDEFINED
}

func renameRequest(values ...*clientsapi.Value) *clientsapi.ActionRequest {
	return &clientsapi.ActionRequest{
		ActionId:  "test-action",
		DeviceId:  "test-device",
		ServiceId: "info",
		Values:    values,
	}
}

func TestInfoNameReadOnlyByDefault(t *testing.T) {
	srv := NewInfo()
	if perms := srv.Name.Perms(); perms != clientsapi.Permissions_PERM_READONLY {
		t.Errorf("expected default name perms PERM_READONLY, got %s", perms)
	}
	if perms := nameAttrPerms(t, srv); perms != clientsapi.Permissions_PERM_READONLY {
		t.Errorf("expected serialized name perms PERM_READONLY, got %s", perms)
	}
}

func TestInfoEnableRename(t *testing.T) {
	srv := NewInfo()
	var got string
	srv.EnableRename(func(newName string) error {
		got = newName
		return nil
	})

	if perms := nameAttrPerms(t, srv); perms != clientsapi.Permissions_PERM_READWRITE {
		t.Errorf("expected serialized name perms PERM_READWRITE after EnableRename, got %s", perms)
	}

	req := renameRequest(&clientsapi.Value{
		Id:   "name",
		Text: &clientsapi.TextValue{Value: "  New Name  "},
	})
	if err := srv.HandleAction(req, nil); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if got != "New Name" {
		t.Errorf("expected handler to receive trimmed %q, got %q", "New Name", got)
	}
}

func TestInfoEnableRenameHandlerError(t *testing.T) {
	srv := NewInfo()
	wantErr := errors.New("backend rejected rename")
	srv.EnableRename(func(newName string) error { return wantErr })

	req := renameRequest(&clientsapi.Value{
		Id:   "name",
		Text: &clientsapi.TextValue{Value: "New Name"},
	})
	if err := srv.HandleAction(req, nil); !errors.Is(err, wantErr) {
		t.Errorf("expected handler error %q, got %v", wantErr, err)
	}
}

func TestInfoEnableRenameRejectsOtherAttributes(t *testing.T) {
	srv := NewInfo()
	srv.EnableRename(func(newName string) error {
		t.Error("handler should not be called")
		return nil
	})

	req := renameRequest(&clientsapi.Value{
		Id:   "model",
		Text: &clientsapi.TextValue{Value: "nope"},
	})
	if err := srv.HandleAction(req, nil); !errors.Is(err, ErrReadOnly) {
		t.Errorf("expected ErrReadOnly, got %v", err)
	}
}

func TestInfoEnableRenameRejectsBadValues(t *testing.T) {
	srv := NewInfo()
	srv.EnableRename(func(newName string) error {
		t.Error("handler should not be called")
		return nil
	})

	if err := srv.HandleAction(renameRequest(&clientsapi.Value{
		Id:   "name",
		Bool: &clientsapi.BoolValue{Value: true},
	}), nil); err == nil {
		t.Error("expected error for non-text value")
	}

	if err := srv.HandleAction(renameRequest(&clientsapi.Value{
		Id:   "name",
		Text: &clientsapi.TextValue{Value: "   "},
	}), nil); err == nil {
		t.Error("expected error for empty name")
	}
}

func TestInfoWriteRejectedWithoutRename(t *testing.T) {
	srv := NewInfo()
	req := renameRequest(&clientsapi.Value{
		Id:   "name",
		Text: &clientsapi.TextValue{Value: "New Name"},
	})
	if err := srv.HandleAction(req, nil); !errors.Is(err, ErrReadOnly) {
		t.Errorf("expected ErrReadOnly for write to read-only name, got %v", err)
	}
}
