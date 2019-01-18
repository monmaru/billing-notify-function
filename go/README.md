
## deploy

```
gcloud functions deploy FUNCTION_NAME --runtime go111 --entry-point F --trigger-resource TRIGGER_BUCKET_NAME --trigger-event google.storage.object.finalize --region asia-northeast1 --project=PROJECT_ID --set-env-vars WEBHOOK=WEBHOOK_URL
```
