import { CurrencyPair, LCDClient, QuotePrice } from "@initia/initia.js";

export async function handleOracle(
    l1lcd: LCDClient,
    filterPairs?: CurrencyPair[],
): Promise<QuotePrice[]> {
    const prices: QuotePrice[] = []
    const pairs = await l1lcd.oracle.currencyPairs();
    const filteredPairs = pairs.filter(
        pair => {
            if (!filterPairs) return true
            return filterPairs.some(
                filterPair => {
                    return (pair.Base == filterPair.Base && pair.Quote == filterPair.Quote)
                }
            )
        })

    for ( const pair of filteredPairs ) {
        const price = await l1lcd.oracle.price(pair)
        prices.push(price)
    }

    return prices
}

async function main() {
    const lcd = new LCDClient('https://lcd.mahalo-1.initia.xyz')
    const res = await handleOracle(lcd,
        [
            new CurrencyPair('BITCOIN','USD')
        ]
    )
    console.log(res)
}

main()