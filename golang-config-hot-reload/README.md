# golang-config-hot-reload

Hot‑reload configuration using Viper + fsnotify with a small HTTP endpoint.

## Run

```bash
go run .
```

## Notes

- Edit `config.yaml` while the app runs and hit `GET /config` to see changes.
