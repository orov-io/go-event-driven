package main

import (
	"log"
	"time"

	"github.com/eapache/go-resiliency/retrier"
)

type User struct {
	Email string
}

type UserRepository interface {
	CreateUserAccount(u User) error
}

type NotificationsClient interface {
	SendNotification(u User) error
}

type NewsletterClient interface {
	AddToNewsletter(u User) error
}

type Handler struct {
	repository          UserRepository
	newsletterClient    NewsletterClient
	notificationsClient NotificationsClient
}

func NewHandler(
	repository UserRepository,
	newsletterClient NewsletterClient,
	notificationsClient NotificationsClient,
) Handler {
	return Handler{
		repository:          repository,
		newsletterClient:    newsletterClient,
		notificationsClient: notificationsClient,
	}
}

func (h Handler) SignUp(u User) error {
	if err := h.repository.CreateUserAccount(u); err != nil {
		return err
	}

	h.AddUserToNewsletterAsync(u)
	h.SendNotificationToUserAsync(u)

	return nil
}

func (h Handler) AddUserToNewsletterAsync(u User) {
	tries := 100
	r := retrier.New(retrier.ConstantBackoff(tries, 100*time.Millisecond), nil)
	go func() {
		err := r.Run(func() error {
			err := h.newsletterClient.AddToNewsletter(u)
			if err != nil {
				log.Printf("failed to add user to the newsletter: %v", err)
			}

			return err
		})

		if err != nil {
			log.Printf("failed to add user after %d tries: %v", tries, err)
		}
	}()
}

func (h Handler) SendNotificationToUserAsync(u User) {
	tries := 100
	r := retrier.New(retrier.ConstantBackoff(tries, 100*time.Millisecond), nil)
	go func() {
		err := r.Run(func() error {
			err := h.notificationsClient.SendNotification(u)
			if err != nil {
				log.Printf("failed to send user notification: %v", err)
			}

			return err
		})

		if err != nil {
			log.Printf("failed to send user notification after %d tries: %v", tries, err)
		}
	}()
}
