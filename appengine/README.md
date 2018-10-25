- Install Google Cloud SDK

- set API_KEY to something in app.yaml

- To run locally

  ```
  dev_appserver.py app.yaml
  ```

  will bring up a local server on http://localhost:8080

- To deploy

  ```
  gcloud projects create PROJECT_ID --set-as-default
  ```

  ```
  gcloud app deploy
  ```

  will deploy on https://PROJECT_ID.appspot.com

- In flowy, set URL and API Key in set storage, and select resync remote storage, which will resync your local items on the web server.

- To get the same items on any other device, open flowy and set storage and API key, and select resync local storage.
