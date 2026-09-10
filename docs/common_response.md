# Common Response

### Backend Json Response:

```json
{
    "code": {{number}},
    "message": {{string}},
    "data": {{object}},
}
```

### Success Code

| Code | Message |
| ---- | ------- |
| 0    | Success |

### Server Error Code

| Code | Message |
| ---- | ------- |
| 500  | Internal Server Error |

### Auth Error Code

| Code | Message |
| ---- | ------- |
| 1001 | Incorrect username or password |
| 1002 | User already exists |
| 1003 | Invalid Param |
| 1004 | Invalid token |