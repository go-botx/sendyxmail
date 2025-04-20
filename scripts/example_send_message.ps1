$token = '111111111111111111111111111111111111111111111111111111111111111111'

$body = @{
    to = 'user@example.com'
    body="Remember to watch new movies!`n`n**Choose yours**"
    buttons = @(
        @(
            @{
                label = "Disney ✨"
                text_align = "left"
                link = "https://disney.com"
                alert_text = "💣💥Warning! Disney remakes can ruin your childhood memories!"
                text_color = "#fbbe4f"
                background_color = "#4a62d8"
                h_size = 5
            },
            @{
                label = "DreamWorks 🌙"
                text_align = "center"
                link = "https://dreamworks.com"
                alert_text = "From 👹Shrek to 🌈Trolls"
                text_color = "#ffffff"
                background_color = "#052c6f"
                h_size = 7
            },
            @{
                label = "🐭 WB 🐈"
                text_align = "right"
                link = "https://www.warnerbros.com/"
                alert_text = "Lego, Tom&Jerry, Looney Tunes, ..."
                text_color = "#000000"
                background_color = "#e1393e"
                h_size = 4
            }
        ),
        @(
            @{
                label = "Illumination 🧞"
                text_align = "center"
                link = "https://www.illumination.com/"
                alert_text = "Minions, MINIONS everywhere!"
                text_color = "#0693e3"
                background_color = "#fcb900"
                h_size = 9
            }
        )
    )
}

$params = @{
    Uri = 'http://127.0.0.1:8000/api/v0/message/with-status'
    Method = 'POST'
    Body =  $body | ConvertTo-Json -Depth 5
    ContentType = 'application/json; charset=utf-8'
    Headers = @{
        Authorization = "Bearer $($token)"
    }
    
}

Invoke-RestMethod @params