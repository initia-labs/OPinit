import { CurrencyPair, LCDClient, QuotePrice } from "@initia/initia.js";

export async function handleOracle(
    l1lcd: LCDClient,
    oraclePairs?: CurrencyPair[],
): Promise<QuotePrice[]> {
    const prices: QuotePrice[] = []
    const pairs = await l1lcd.oracle.currencyPairs();
    const filteredPairs = pairs.filter(
        pair => {
            if (!oraclePairs) return true
            return oraclePairs.some(
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

export function toCurrencyPair(
    pairs: string
): CurrencyPair[] {
    return pairs.split(',').map(
        pair => {
            const [base, quote] = pair.split('/')
            return new CurrencyPair(base, quote)
        }
    )
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