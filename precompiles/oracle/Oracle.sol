// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

// define the precompiled address
address constant ORACLE_PRECOMPILE_ADDRESS = 0x0000000000000000000000000000000000001008;

IOracle constant ORACLE_CONTRACT = IOracle(ORACLE_PRECOMPILE_ADDRESS);

// create interface
interface IOracle {
    // getExchangeRates queries the current exchange rates on the module
    function getExchangeRates()
        external
        view
        returns (DenomOracleExchangeRate[] memory);

    // getOracleTwaps queries the module's twap withing a lookback period
    function getOracleTwaps(
        uint256 lookback_seconds
    ) external view returns (OracleTwap[] memory);

    // getActives queries the active assets list on the module
    function getActives() external view returns (string[] memory);

    // getPriceSnapshotHistory queries the price history with snapshots
    function getPriceSnapshotHistory()
        external
        view
        returns (PriceSnapshot[] memory);

    // getFeederDelegation queries the feeder delegated based on the validator address
    function getFeederDelegation(
        string memory validatorAddress
    ) external view returns (string memory);

    // getVotePenaltyCounter queries the vote penalty counters based on the validator address
    function getVotePenaltyCounter(
        string memory validatorAddress
    ) external view returns (VotePenaltyCounter memory);

    // OracleExchangeRate represents the information associated to a denom in a
    // exchange rate
    struct OracleExchangeRate {
        string exchangeRate;
        string lastUpdate;
        uint256 lastUpdateTimestamp;
    }

    // DenomOracleExchangeRate represents a exchange rate on the module
    struct DenomOracleExchangeRate {
        string denom;
        OracleExchangeRate oracleExchangeRate;
    }

    // OracleTwap represents the twap output from the module
    struct OracleTwap {
        string denom;
        string twap;
        uint256 lookbackSeconds;
    }

    // PriceSnapshot represents an snapshot
    struct PriceSnapshot {
        uint256 snapshotTimestamp;
        DenomOracleExchangeRate[] PriceSnapshotItems;
    }

    // VotePenaltyCounter represents the votepenalty result from module
    struct VotePenaltyCounter {
        uint256 missCount;
        uint256 abstainCount;
        uint256 successCount;
    }
}
