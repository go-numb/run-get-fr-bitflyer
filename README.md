# run-get-fr-bitflyer
指定スケジュールでTicker,FRを取得するプログラム  

## endpoint
- https://<deployed>/api/get-ticker?product_code=<ProductCode>

## Usage
- Google cloud schedules(set header: ProjectId)
- Google cloud run
- Google cloud firestore

## Envs
- PORT
- PROJECTID