// Post a Tweet from your Pipeline
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/michimani/gotwi"
	"github.com/michimani/gotwi/tweet/managetweet"
	"github.com/michimani/gotwi/tweet/managetweet/types"

	"dagger/twitter/internal/dagger"
)

type Twitter struct {
	// +private
	Auth Auth
	// +private
	Message Message
}

type Message struct {
	DirectMessageDeepLink string
	ForSuperFollowersOnly bool
	GeoPlaceID            string
	MediaIDs              []string
	TaggedUserID          string
	PollOptions           []string
	PollDurationMinutes   int
	QuoteTweetID          string
	ExcludedUsersIDsReply []string
	InReplyToTweetId      string
	ReplySettings         string
	Text                  string
}

type Auth struct {
	// +private
	ConsumerKey *dagger.Secret

	// +private
	ConsumerSecret *dagger.Secret

	// +private
	OauthToken *dagger.Secret

	// +private
	OauthSecret *dagger.Secret
}

func New(
	ctx context.Context,
	// Twitter consumer key
	// +required
	consumerKey *dagger.Secret,

	// Twitter consumer secret
	// +required
	consumerSecret *dagger.Secret,

	// Twitter Oauth token
	// +required
	oauthToken *dagger.Secret,

	// Twitter Oauth secret
	// +required
	oauthSecret *dagger.Secret,
) *Twitter {
	return &Twitter{
		Auth: Auth{
			ConsumerKey:    consumerKey,
			ConsumerSecret: consumerSecret,
			OauthToken:     oauthToken,
			OauthSecret:    oauthSecret,
		},
	}
}

// Creates a post to Twitter/X
//
// example usage from the CLI
// ex.: dagger call --oauth-token env:<OAUTH-TOKEN> --oauth-secret env:<OAUTH-SECRET> send-tweet --text "Example text"
func (t *Twitter) SendTweet(
	ctx context.Context,
	// Message you want to tweet
	// +optional
	text string,
	// Tweets a link directly to a Direct Message conversation with an account
	// Example: "https://twitter.com/messages/compose?recipient_id=2244994945"
	// +optional
	directMessageLink string,
	// Allows you to Tweet exclusively for Super Followers
	// +optional
	superUserOnly bool,
	// Place ID being attached to the Tweet for geo location
	// +optional
	geoPlaceID string,
	// A list of Media IDs being attached to the Tweet
	// +optional
	mediaIDs []string,
	// A list of User IDs being tagged in the Tweet with Media. If the user you're tagging doesn't have photo-tagging enabled, their names won't show up in the list of tagged users even though the Tweet is successfully created
	// +optional
	taggedUserID string,
	// A list of poll options for a Tweet with a poll
	// +optional
	pollOptions []string,
	// Duration of the poll in minutes for a Tweet with a poll
	// +optional
	pollDurationMinutes int,
	// Link to the Tweet being quoted
	// +optional
	quotedTweetID string,
	// A list of User IDs to be excluded from the reply Tweet thus removing a user from a thread
	// +optional
	excludedUsersIDsReply []string,
	// Tweet ID of the Tweet being replied to
	// +optional
	inReplytoTweetId string,
	// Settings to indicate who can reply to the Tweet. Options include "mentionedUsers" and "following". If the field isn’t specified, it will default to everyone
	// +optional
	repplySettings string,
) (string, error) {
	message := t.Message.createMessage(
		text,
		directMessageLink,
		superUserOnly,
		geoPlaceID,
		mediaIDs,
		taggedUserID,
		pollOptions,
		pollDurationMinutes,
		quotedTweetID,
		excludedUsersIDsReply,
		inReplytoTweetId,
		repplySettings,
	)

	post := &types.CreateInput{}

	if message.Text != "" {
		post.Text = strPtr(message.Text)
	}

	if message.DirectMessageDeepLink != "" {
		post.DirectMessageDeepLink = strPtr(message.DirectMessageDeepLink)
	}

	if message.ForSuperFollowersOnly {
		post.ForSuperFollowersOnly = boolPtr(message.ForSuperFollowersOnly)
	}

	if message.GeoPlaceID != "" {
		post.Geo = &types.CreateInputGeo{PlaceID: strPtr(message.GeoPlaceID)}
	}

	if message.MediaIDs != nil || message.TaggedUserID != "" {
		post.Media = &types.CreateInputMedia{
			MediaIDs:     message.MediaIDs,
			TaggedUserID: strPtr(message.TaggedUserID),
		}
	}

	if message.PollDurationMinutes > 0 && message.PollOptions != nil {
		post.Poll = &types.CreateInputPoll{
			Options:         message.PollOptions,
			DurationMinutes: intPtr(message.PollDurationMinutes),
		}
	}

	if message.QuoteTweetID != "" {
		post.QuoteTweetID = strPtr(message.QuoteTweetID)
	}

	if message.InReplyToTweetId != "" || message.ExcludedUsersIDsReply != nil {
		post.Reply = &types.CreateInputReply{
			InReplyToTweetID:    message.InReplyToTweetId,
			ExcludeReplyUserIDs: message.ExcludedUsersIDsReply,
		}
	}

	if message.ReplySettings != "" {
		post.ReplySettings = strPtr(message.ReplySettings)
	}

	oauthTokenString, err := t.Auth.OauthToken.Plaintext(ctx)
	if err != nil {
		return "", err
	}

	oauthTokenSecretString, err := t.Auth.OauthSecret.Plaintext(ctx)
	if err != nil {
		return "", err
	}

	consumerKeyString, err := t.Auth.ConsumerKey.Plaintext(ctx)
	if err != nil {
		return "", err
	}

	consumerSecretString, err := t.Auth.ConsumerSecret.Plaintext(ctx)
	if err != nil {
		return "", err
	}

	client, err := createClientConnection(
		oauthTokenString,
		oauthTokenSecretString,
		consumerKeyString,
		consumerSecretString,
	)
	if err != nil {
		return "", err
	}

	response, err := managetweet.Create(ctx, client, post)
	if err != nil {
		return "", err
	}

	result := fmt.Sprintln(
		"[%s]: %s",
		gotwi.StringValue(response.Data.ID),
		gotwi.StringValue(response.Data.Text),
	)
	return result, nil
}

func (m *Message) createMessage(
	text string,
	directMessageLink string,
	superUserOnly bool,
	geoPlaceID string,
	mediaIDs []string,
	taggedUserID string,
	pollOptions []string,
	pollDurationMinutes int,
	quotedTweetID string,
	excludedUsersIDsReply []string,
	inReplytoTweetId string,
	repplySettings string,
) *Message {
	if text != "" {
		m.Text = text
	}

	if directMessageLink != "" {
		m.DirectMessageDeepLink = directMessageLink
	}

	if superUserOnly {
		m.ForSuperFollowersOnly = superUserOnly
	}

	if geoPlaceID != "" {
		m.GeoPlaceID = geoPlaceID
	}

	if taggedUserID != "" {
		if mediaIDs != nil {
			m.TaggedUserID = taggedUserID
			m.MediaIDs = mediaIDs
		}
		m.TaggedUserID = taggedUserID
	}

	if pollDurationMinutes >= 0 && pollOptions != nil {
		m.PollDurationMinutes = pollDurationMinutes
		m.PollOptions = pollOptions
	}

	if quotedTweetID != "" {
		m.QuoteTweetID = quotedTweetID
	}

	if inReplytoTweetId != "" {
		if excludedUsersIDsReply != nil {
			m.ExcludedUsersIDsReply = excludedUsersIDsReply
			m.InReplyToTweetId = inReplytoTweetId
		}
		m.InReplyToTweetId = inReplytoTweetId
	}

	if repplySettings != "" {
		m.ReplySettings = repplySettings
	}

	return m
}

// Delete a Tweet with a tweetID
//
// example usage from the CLI
// ex.: dagger call --oauth-token env:<OAUTH-TOKEN> --oauth-secret env:<OAUTH-SECRET> delete-tweet --tweet-id "123456789"
func (t *Twitter) DeleteTweet(
	ctx context.Context,
	// The Tweet ID you are deleting
	// +requires
	tweetID string,
) (bool, error) {
	oauthTokenString, err := t.Auth.OauthToken.Plaintext(ctx)
	if err != nil {
		return false, err
	}

	oauthTokenSecretString, err := t.Auth.OauthSecret.Plaintext(ctx)
	if err != nil {
		return false, err
	}

	consumerKeyString, err := t.Auth.ConsumerKey.Plaintext(ctx)
	if err != nil {
		return false, err
	}

	consumerSecretString, err := t.Auth.ConsumerSecret.Plaintext(ctx)
	if err != nil {
		return false, err
	}

	client, err := createClientConnection(
		oauthTokenString,
		oauthTokenSecretString,
		consumerKeyString,
		consumerSecretString,
	)
	if err != nil {
		return false, err
	}

	postToDelete := &types.DeleteInput{
		ID: tweetID,
	}

	result, err := managetweet.Delete(ctx, client, postToDelete)
	if err != nil {
		return false, err
	}

	boolResult := gotwi.BoolValue(result.Data.Deleted)

	return boolResult, nil
}

func createClientConnection(
	oauthTokenString, oauthTokenSecretString, consumerKeyString, consumerSecretString string,
) (*gotwi.Client, error) {
	if token := os.Setenv("GOTWI_API_KEY", oauthTokenString); token != nil {
		return nil, token
	}

	if secret := os.Setenv("GOTWI_API_KEY_SECRET", oauthTokenSecretString); secret != nil {
		return nil, secret
	}

	if token := os.Setenv("GOTWI_ACCESS_TOKEN", consumerKeyString); token != nil {
		return nil, token
	}

	if secret := os.Setenv("GOTWI_ACCESS_TOKEN_SECRET", consumerSecretString); secret != nil {
		return nil, secret
	}

	in := &gotwi.NewClientInput{
		AuthenticationMethod: gotwi.AuthenMethodOAuth1UserContext,
		OAuthToken:           consumerKeyString,
		OAuthTokenSecret:     consumerSecretString,
		Debug:                true,
	}

	client, err := gotwi.NewClient(in)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return gotwi.String(s)
}

func boolPtr(b bool) *bool {
	return gotwi.Bool(b)
}

func intPtr(i int) *int {
	if i == 0 {
		return nil
	}
	return &i
}
