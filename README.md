# git-helper

Helps facilitate git-based operations with Kubefirst's platform.

## Components

The tool offers the ability to create and delete repository/project webhooks along with other features.

Data can be passed in as arguments or retrieved from Secrets.

### `ngrok` Sync

A specific use case for this tool is assisting with automating refreshing `ngrok` tunnels and updating webhooks with updated URLs.

The logic is as follows:

- ngrok starts
- ngrok api is available
- sync app calls the api to get the new tunnel endpoint url
- sync app reads atlantis webhook token from existing atlantis secret
- write new webhook with updated tunnel endpoint, delete the old one
- cleanup webhook when platform is destroyed
