FROM golang:latest

# Устанавливаем git для клонирования репозитория
RUN apt-get update && apt-get install -y git && apt-get install -y nano

# Клонируем репозиторий
RUN git clone https://github.com/Nefar10/Afina_go.git

# Указываем рабочий каталог
WORKDIR ./Afina_go

# Сборка приложения
RUN go build -o main

#Пробрасываем текущую папку внутрь крнтейнера
VOLUME /data

# Установка команды запуска приложения по умолчанию
CMD ["./main"]


ENV AFINA_API_KEY=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
ENV TB_API_KEY=6254972536:AAE-5w19NlVYphiJKbUaelr244oDiaSU184
ENV AFINA_NAMES=Адам,Adam
ENV AFINA_GENDER=Male