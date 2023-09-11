import type { Config } from '@jest/types';

// Jest configuration
const config: Config.InitialOptions = {
  roots: ['<rootDir>/src', '<rootDir>/src/test'],
  transform: {
    '^.+\\.ts?$': 'ts-jest',
  },
  testRegex: '(/__tests__/.*|(\\.|/)(test|spec))\\.ts?$',
  moduleFileExtensions: ['ts', 'js', 'json', 'node'],
  moduleNameMapper: {
    '^graphql$': '<rootDir>/node_modules/graphql',
    '^orm/(.*)$': '<rootDir>/src/orm/$1',
  },
  modulePaths: ['<rootDir>/src', '<rootDir>/lib'],
  moduleDirectories: ['node_modules'],
  // collectCoverage: true,
  collectCoverageFrom: ['src/**/*.{ts,js}'],
  coverageDirectory: 'coverage',
  testEnvironment: 'node',
  globals: {
    'ts-jest': {
      tsconfig: 'tsconfig.json',
    },
  },
};

export default config;
