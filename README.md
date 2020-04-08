# Dynamic Multitwitch via Cloudflare

Updates a redirect page rule to only include channels that are currently live based on a given list.
Defaults to all channels if none are online.

## Dependencies

Relies on calling to a [Twitch Redis Cache](https://github.com/UpDownLeftDie/twitch-redis-cache) server to check the status of the streams

## Instructions

Update `config.json` with your cloudflare credentials and the Zone ID and the ID for the page rule you want to update
Update `usernames.json` with like of usernames you want to check
