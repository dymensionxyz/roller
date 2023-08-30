# generate a random number between 1 and 10,000
RANDOM_NUMBER=$((1 + RANDOM % 10000))

/usr/local/bin/roller config init loadtest_$RANDOM_NUMBER-1