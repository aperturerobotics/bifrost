package spacewave_chat

import spacewave_chat_content "github.com/s4wave/spacewave/sdk/chat/content"

// ChatMessageContent is the shared generated content type used by storage and RPC.
type ChatMessageContent = spacewave_chat_content.ChatMessageContent

// ChatMessageContent_Text retains the public text-content constructor.
type ChatMessageContent_Text = spacewave_chat_content.ChatMessageContent_Text

// ChatMessageContent_Ciphertext carries an endpoint-encrypted message envelope.
type ChatMessageContent_Ciphertext = spacewave_chat_content.ChatMessageContent_Ciphertext

// ChatCiphertext is the shared encrypted envelope used by storage and RPC.
type ChatCiphertext = spacewave_chat_content.ChatCiphertext

// ChatRelation is the shared message relationship used by encrypted transports.
type ChatRelation = spacewave_chat_content.ChatRelation

// ChatAnnotation is an attributed public reaction to a channel message.
type ChatAnnotation = spacewave_chat_content.ChatAnnotation

// ChatMessageContent_Annotation carries a public reaction independently of message encryption.
type ChatMessageContent_Annotation = spacewave_chat_content.ChatMessageContent_Annotation
