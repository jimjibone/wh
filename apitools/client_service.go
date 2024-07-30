package apitools

import (
	"fmt"

	clientsapi "github.com/jimjibone/woodhouse-4/api/go/v1/clients"
)

func ImageResponseString(res *clientsapi.ImageResponse) string {
	return fmt.Sprintf("id:%q, status:%s, details:%q, data:%d", res.GetRequestId(), res.GetStatus(), res.GetDetails(), len(res.GetData()))
}
