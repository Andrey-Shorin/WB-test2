start:
docker compose build
 docker compose up 
 curl http://localhost:8080/cities
 выведет список доступных городов например {"id":4,"name":"London","country":"GB","lat":51.507320404052734,"lon":-0.12764739990234375}

 теперь можно посмотреть доступные даты для города по его id
 http://localhost:8080/cities/4/forecast
это вернет список доступных дат в виде timestamp
{"name":"London","country":"GB","average_temp":289.571282051282,"available_dates":["1720634400","1720645200", ... ,"1721034000","1721044800"]}

можно выбрать конкретную дату 
curl http://localhost:8080/cities/4/forecast/1720634400

и получить всю доступную информацию

