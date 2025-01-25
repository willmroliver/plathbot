source .env

TAGS=""

# account emoji reddit stats pfp games

if [ $API_ACCOUNT ] ; then 
    TAGS="$TAGS account"
fi

if [ $API_EMOJI ] ; then 
    TAGS="$TAGS emoji"
fi

if [ $API_REDDIT ] ; then 
    TAGS="$TAGS reddit"
fi

if [ $API_STATS ] ; then 
    TAGS="$TAGS stats"
fi

if [ $API_PFP ] ; then 
    TAGS="$TAGS pfp"
fi

if [ $API_GAMES ] ; then 
    TAGS="$TAGS games"
fi

go run -tags="$TAGS" src/main.go
