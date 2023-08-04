import { logger } from 'lib/logger';
import config from 'config';
import { Monitor } from './Monitor';
import { getL2Denom } from 'lib/util';
import { ChallengerOutputEntity, DepositTxEntity } from 'orm';

export class L1Monitor extends Monitor {
  public name(): string {
    return 'challenger_l1_monitor';
  }

  public color(): string {
    return 'blue';
  }

  public async handleEvents(): Promise<void> {
    const searchRes = await config.l1lcd.tx.search({
      events: [{ key: 'tx.height', value: (this.syncedHeight + 1).toString() }]
    });
    const events = searchRes.txs
      .flatMap((tx) => tx.logs ?? [])
      .flatMap((log) => log.events);

    const lastOutput = await this.getLastOutputFromDB();
    const lastIndex = lastOutput.length == 0 ? -1 : lastOutput[0].outputIndex;

    for (const evt of events) {
      if (evt.type !== 'move') continue;

      const attrMap: { [key: string]: string } = evt.attributes.reduce(
        (obj, attr) => {
          obj[attr.key] = attr.value;
          return obj;
        },
        {}
      );

      const data: { [key: string]: string } = JSON.parse(attrMap['data']);
      if (data['l2_id'] !== config.L2ID) continue;
      if (attrMap['type_tag'] !== '0x1::op_bridge::TokenBridgeInitiatedEvent')
        continue;

      // handle token bridge initiated event
      const denom = getL2Denom(Buffer.from(data['l2_token']));

      const entity: DepositTxEntity = {
        sequence: Number.parseInt(data['l1_sequence']),
        sender: data['from'],
        receiver: data['to'],
        amount: Number.parseInt(data['amount']),
        outputIndex: lastIndex + 1,
        finalizedOutputIndex: null,
        coinType: denom,
        isChecked: false
      };

      await this.db.getRepository(DepositTxEntity).save(entity);
    }
  }
}
