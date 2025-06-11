source .env

TAGS=""

# account emoji reddit stats pfp games

if [ "$API_ACCOUNT" -eq 1 ] ; then 
    TAGS="$TAGS account"
fi

if [ "$API_EMOJI" -eq 1 ] ; then 
    TAGS="$TAGS emoji"
fi

if [ "$API_REDDIT" -eq 1 ] ; then 
    TAGS="$TAGS reddit"
fi

if [ "$API_STATS" -eq 1 ] ; then 
    TAGS="$TAGS stats"
fi

if [ "$API_PFP" -eq 1 ] ; then 
    TAGS="$TAGS pfp"
fi

if [ "$API_GAMES" -eq 1 ] ; then 
    TAGS="$TAGS games"
fi

go build -v -mod=vendor -tags="$TAGS" ./src/main.go
