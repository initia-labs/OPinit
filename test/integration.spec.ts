import { PostgreSqlContainer } from '@testcontainers/postgresql';
import {GenericContainer} from 'testcontainers';
import { Client } from 'pg';
import {
    ConnectionOptionsReader,
    DataSource,
    DataSourceOptions
} from 'typeorm';
import { initORM } from 'worker/bridgeExecutor/db';



describe('PostgreSQL Integration Tests', () => {
    jest.setTimeout(100000);
    let container;
    let client;

    beforeAll(async () => {
        container = await new GenericContainer('postgres')
            .withEnvironment({
                POSTGRES_USER: 'jungsuhan',
                POSTGRES_PASSWORD: 'jungsuhan',
                POSTGRES_DB: 'challenger',
            })
            .withExposedPorts(5432)
            .start();
            
        const pgOpts = [
            {

            }
        ]

        // use initORM() to connect to the database
        

        // await client.connect();
    });
  
    afterAll(async () => {
        // await client.end();
        await container.stop();
    });
  
    test('should connect to PostgreSQL', async () => {
        await initORM();
        console.log("hh")
        // const result = await client.query('SELECT 1 AS value');
        // expect(result.rows[0].value).toBe(1);
    });
});
