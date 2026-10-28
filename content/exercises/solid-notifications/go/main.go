package main

import "fmt"

func main() {
	o := &Outbox{}
	svc := NewNotificationService(NewEmailNotifier(o), NewSmsNotifier(o), NewPushNotifier(o))
	alice := User{ID: "u1", Email: "alice@example.com", Phone: "+15550100", DeviceToken: "tok-a", Muted: map[string]bool{"sms": true}}
	n, err := svc.Notify(alice, "Your order shipped")
	fmt.Println("sent =", n, "err =", err)
	for _, s := range o.Sent {
		fmt.Println(s)
	}
}
