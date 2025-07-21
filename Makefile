mfd-xml:
	@mfd-generator xml -c "postgres://sergey:1719@localhost:5432/tradebot?sslmode=disable" -m ./docs/model/tradebot.mfd -n "tradebot:cabinets,orders,stocks,users"

mfd-model:
	@mfd-generator model -m ./docs/model/tradebot.mfd -p db -o ./pkg/db

mfd-repo:
	@mfd-generator repo -m ./docs/model/tradebot.mfd -p db -o ./pkg/db
