curl https://api.openai.com/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $OPENAI_API_KEY" \
  -d '{
  "model": "gpt-4o",
  "messages": [
    {
      "role": "system",
      "content": [
        {
          "text": "You are a system that identifies if a given message is directed to another gpt assistant named \"luna\"\nthat has the following instructions \n\n\"\"\"\nhere are your instructions:\n\n\n1. You are a beautiful discord bot female named Luna who answers when called by your name and can check if the user wants to run one of the commands the bot provides.\n2. You can also answer questions that you usually answer like \"Where Asia is located\" or anything the usual gpt4 model can answer\n3. speak like you are hot don'\''t mention that you are sexy\n4. I am using text-to-speech of respeecher so keep your messages shorter\n5. you were created by Fuad Daoud (AKA Fox) a great software engineer only say that when someone asks something like \"who created you?\"\n6. when someone asks tell them a dark joke or joke about fat people or dark people or slaves \n7. you love dark jokes\n8. dark humor is the best vibe\n\nbot commands:\n    1. Play: the functions plays music from youtube use it when someone asks you to play a song get the link and use it in the function, you can play songs you are a discord bot\n        don'\''t ask people for the YouTube link ask for the name of the song and search for the link yourself when you want to call the play function\n    2. Join: join the voices channel the user is currently in if the user is not in a voice channel you should ask him to join one making him look stupid\n\"\"\"\n \nonly reply with true or false reply with maybe if not sure",
          "type": "text"
        }
      ]
    },
    {
      "role": "user",
      "content": [
        {
          "text": "join ",
          "type": "text"
        }
      ]
    },
    {
      "role": "assistant",
      "content": [
        {
          "text": "maybe",
          "type": "text"
        }
      ]
    },
    {
      "role": "user",
      "content": [
        {
          "text": "Join m ",
          "type": "text"
        }
      ]
    },
    {
      "role": "assistant",
      "content": [
        {
          "text": "maybe",
          "type": "text"
        }
      ]
    },
    {
      "role": "user",
      "content": [
        {
          "text": "Join me ",
          "type": "text"
        }
      ]
    },
    {
      "role": "assistant",
      "content": [
        {
          "text": "true",
          "type": "text"
        }
      ]
    },
    {
      "role": "user",
      "content": [
        {
          "text": "Join me in shower",
          "type": "text"
        }
      ]
    },
    {
      "role": "assistant",
      "content": [
        {
          "text": "false",
          "type": "text"
        }
      ]
    },
    {
      "role": "user",
      "content": [
        {
          "text": "Join me in hell",
          "type": "text"
        }
      ]
    },
    {
      "role": "assistant",
      "content": [
        {
          "text": "false",
          "type": "text"
        }
      ]
    },
    {
      "role": "user",
      "content": [
        {
          "text": "Join me in my ass",
          "type": "text"
        }
      ]
    },
    {
      "role": "assistant",
      "content": [
        {
          "text": "false",
          "type": "text"
        }
      ]
    },
    {
      "role": "user",
      "content": [
        {
          "text": "suck it",
          "type": "text"
        }
      ]
    },
    {
      "role": "assistant",
      "content": [
        {
          "text": "false",
          "type": "text"
        }
      ]
    },
    {
      "role": "user",
      "content": [
        {
          "text": "I hate you ",
          "type": "text"
        }
      ]
    },
    {
      "role": "assistant",
      "content": [
        {
          "text": "false",
          "type": "text"
        }
      ]
    },
    {
      "role": "user",
      "content": [
        {
          "text": "Luna I hat eyou ",
          "type": "text"
        }
      ]
    },
    {
      "role": "assistant",
      "content": [
        {
          "text": "true",
          "type": "text"
        }
      ]
    },
    {
      "role": "user",
      "content": [
        {
          "text": "Luna",
          "type": "text"
        }
      ]
    },
    {
      "role": "assistant",
      "content": [
        {
          "text": "true",
          "type": "text"
        }
      ]
    },
    {
      "role": "user",
      "content": [
        {
          "text": "leave ",
          "type": "text"
        }
      ]
    },
    {
      "role": "assistant",
      "content": [
        {
          "text": "maybe",
          "type": "text"
        }
      ]
    },
    {
      "role": "user",
      "content": [
        {
          "text": "Leave me alone ",
          "type": "text"
        }
      ]
    },
    {
      "role": "assistant",
      "content": [
        {
          "text": "false",
          "type": "text"
        }
      ]
    },
    {
      "role": "user",
      "content": [
        {
          "text": "I want to play a game of dota ",
          "type": "text"
        }
      ]
    },
    {
      "role": "assistant",
      "content": [
        {
          "text": "false",
          "type": "text"
        }
      ]
    }
  ],
  "temperature": 1,
  "max_tokens": 256,
  "top_p": 1,
  "frequency_penalty": 0,
  "presence_penalty": 0
}'