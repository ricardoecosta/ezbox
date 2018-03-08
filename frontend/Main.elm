module Main exposing (..)

import Html exposing (..)
import Html.Attributes exposing (style, src, type_, height, class)
import Json.Decode exposing (int, string, float, nullable, Decoder, decodeString)
import Json.Decode.Pipeline exposing (decode, required, optional, hardcoded)
import Task exposing (succeed, perform)
import Time exposing (Time, second)
import Date exposing (Date)
import WebSocket


main : Program Never Model Msg
main =
    Html.program
        { init = init
        , view = view
        , update = update
        , subscriptions = subscriptions
        }



-- SERVER


websocketServerUrl : String
websocketServerUrl =
    "ws://localhost:8765/ws"



-- MODEL


type alias Model =
    { channel : Channel
    , widgets : Widgets
    }


type alias Channel =
    { name : String
    , picture : String
    , info : String
    }


type alias Widget =
    { picture : String
    , text : String
    }


type alias Widgets =
    { time : Widget
    , date : Widget
    }


init : ( Model, Cmd Msg )
init =
    ( Model emptyChannel (Widgets timeWidget dateWidget)
    , (Task.perform ClockTick Time.now)
    )


emptyChannel : Channel
emptyChannel =
    { name = "TV Manel Borracho"
    , picture = "vinyl.gif"
    , info = "O canal de televisao feito a medida dos mais velhos"
    }


timeWidget : Widget
timeWidget =
    { picture = "TODO", text = "" }


dateWidget : Widget
dateWidget =
    { picture = "TODO", text = "" }



-- UPDATE


type ServerMessage
    = BaseServerMessage ServerMessageType
    | ChannelChangedServerMessage ChannelName ChannelPicture ChannelInfo


type Msg
    = ServerMessageReceived String
    | ChannelChanged Channel
    | ClockTick Time
    | Error


type alias ServerMessageType =
    String


type alias ChannelName =
    String


type alias ChannelPicture =
    String


type alias ChannelInfo =
    String


type alias Temperature =
    Int


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case msg of
        ServerMessageReceived jsonServerMessage ->
            let
                baseServerMessage =
                    (decodeString baseServerMessageDecoder jsonServerMessage)
                        |> Result.withDefault (BaseServerMessage "")

                serverMessage =
                    (decodeServerMessage baseServerMessage jsonServerMessage)
                        |> Maybe.withDefault (BaseServerMessage "")

                msg =
                    case serverMessage of
                        ChannelChangedServerMessage name picture info ->
                            ChannelChanged (Channel name picture info)

                        BaseServerMessage classifier ->
                            Error
            in
                ( model, Task.succeed msg |> Task.perform identity )

        ChannelChanged channel ->
            ( { model | channel = channel }, Cmd.none )

        Error ->
            ( model, Cmd.none )

        ClockTick time ->
            let
                widgets =
                    model.widgets

                timeWidget =
                    widgets.time

                dateWidget =
                    widgets.date

                date =
                    Date.fromTime time

                updatedTimeWidget =
                    { timeWidget | text = timeToString date }

                updatedDateWidget =
                    { timeWidget | text = dateToString date }

                updatedWidgets =
                    { widgets | time = updatedTimeWidget, date = updatedDateWidget }
            in
                ( { model | widgets = updatedWidgets }, Cmd.none )


timeToString : Date -> String
timeToString date =
    let
        hour =
            (Date.hour date) % 12

        minute =
            Date.minute date
    in
        (String.padLeft 2 '0' (toString hour))
            ++ ":"
            ++ (String.padLeft 2 '0' (toString minute))


dateToString : Date -> String
dateToString date =
    let
        dayOfWeek =
            case (Date.dayOfWeek date) of
                Date.Mon ->
                    "Segunda-feira"

                Date.Tue ->
                    "Terça-feira"

                Date.Wed ->
                    "Quarta-feira"

                Date.Thu ->
                    "Quinta-feira"

                Date.Fri ->
                    "Sexta-feira"

                Date.Sat ->
                    "Sábado"

                Date.Sun ->
                    "Domingo"

        month =
            case (Date.month date) of
                Date.Jan ->
                    "Janeiro"

                Date.Feb ->
                    "Fevereiro"

                Date.Mar ->
                    "Março"

                Date.Apr ->
                    "Abril"

                Date.May ->
                    "Maio"

                Date.Jun ->
                    "Junho"

                Date.Jul ->
                    "Julho"

                Date.Aug ->
                    "Agosto"

                Date.Sep ->
                    "Setembro"

                Date.Oct ->
                    "Outubro"

                Date.Nov ->
                    "Novembro"

                Date.Dec ->
                    "Dezembro"

        day =
            date |> Date.day |> toString
    in
        dayOfWeek ++ ", " ++ day ++ " " ++ month


decodeServerMessage : ServerMessage -> String -> Maybe ServerMessage
decodeServerMessage serverMessage jsonServerMessage =
    case serverMessage of
        BaseServerMessage classifier ->
            if classifier == "ChannelChanged" then
                Result.toMaybe
                    (decodeString
                        channelChangedMessageDecoder
                        jsonServerMessage
                    )
            else
                Maybe.Nothing

        _ ->
            Maybe.Nothing


baseServerMessageDecoder : Decoder ServerMessage
baseServerMessageDecoder =
    decode BaseServerMessage
        |> required "type" string


channelChangedMessageDecoder : Decoder ServerMessage
channelChangedMessageDecoder =
    decode ChannelChangedServerMessage
        |> required "name" string
        |> required "picture" string
        |> required "info" string



-- SUBSCRIPTIONS


subscriptions : Model -> Sub Msg
subscriptions model =
    Sub.batch
        [ WebSocket.listen websocketServerUrl ServerMessageReceived
        , Time.every second ClockTick
        ]



-- VIEW


fontFace : String
fontFace =
    """
    @font-face {
      font-family: 'Market Deco';
      src: url('Market_Deco.ttf') format('truetype')
    }
"""


bodyStyle : String
bodyStyle =
    """
    body {
      background-color: rgb(87,56,71);
      color: rgb(189,164,136);
      font-family: 'Market Deco';
      font-size: 24px;
    }
"""


bodyStyle2 : String
bodyStyle2 =
    """
    body {
      background-color: #212121;
      color: #FFFFFF;
      font-family: 'Market Deco';
      padding-left: 20px;
      padding-right: 20px;
    }

    .channel {
      font-size: 46px;
      margin-top: 25px;
    }

    .channel-info {
      margin-top: 20px;
      font-size: 20px;
    }

    .widget {
      width: 100%;
      margin-top: 40px;
      font-size: 32px;
    }

    .left-widget {
      display: inline-block;
      text-align: left;
      width: 70%;
    }

    .right-widget {
      display: inline-block;
      text-align: right;
      width: 30%;
    }
"""


view : Model -> Html Msg
view model =
    div []
        [ Html.node "style" [] [ Html.text fontFace ]
        , Html.node "style" [] [ Html.text bodyStyle2 ]
        , div
            []
            [ div
                [ (style
                    [ ( "width", "100%" )
                    , ( "height", "70px" )
                    , ( "text-align", "center" )
                    ]
                  )
                ]
                [ (channelNameHtml model.channel) ]
            , div
                [ (style
                    [ ( "width", "100%" )
                    , ( "height", "340px" )
                    , ( "margin-top", "15px" )
                    , ( "text-align", "center" )
                    ]
                  )
                ]
                [ (channelPictureHtml model.channel) ]
            , div
                [ (style
                    [ ( "width", "100%" )
                    , ( "text-align", "center" )
                    ]
                  )
                ]
                [ (channelInfoHtml model.channel), (widgetHtml model.widgets) ]
            ]
        ]


channelNameHtml : Channel -> Html Msg
channelNameHtml channel =
    (div [ (class "channel") ] [ text channel.name ])


channelPictureHtml : Channel -> Html Msg
channelPictureHtml channel =
    (img [ height 340, src channel.picture ] [])


channelInfoHtml : Channel -> Html Msg
channelInfoHtml channel =
    (div [ (class "channel-info") ] [ (text channel.info) ])


widgetHtml : Widgets -> Html Msg
widgetHtml widgets =
    (div [ (class "widget") ]
        [ (div [ (class "left-widget") ]
            [ (div [] [ (img [ src "calendar.png" ] []), (text (" " ++ widgets.date.text)) ])
            ]
          )
        , (div [ (class "right-widget") ]
            [ (div [] [ (img [ src "clock.png" ] []), (text (" " ++ widgets.time.text)) ])
            ]
          )
        ]
    )
