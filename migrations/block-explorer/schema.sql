CREATE TABLE chain_info (
                            chain_id                            TEXT    NOT NULL,
                            "name"                              TEXT    NOT NULL,
                            chain_type                          TEXT    NOT NULL, -- "evm", "wasm", "cosmos" or empty. Currently only the first 2 values are available
                            bech32                              JSONB   NOT NULL,
                            denoms                              JSONB   NOT NULL,
                            be_json_rpc_urls                    TEXT[], -- sorted, the best one is the first, might be empty
                            latest_indexed_block                BIGINT  NOT NULL DEFAULT 0, -- the latest successfully indexed block height
                            increased_latest_indexed_block_at   BIGINT NOT NULL DEFAULT 0, -- the epoch UTC seconds when the latest_indexed_block updated with greater value
                            postponed                           BOOLEAN, -- true if the chain is postponed/stopped operation
                            keep_recent_account_tx_count        INT, -- number of recent account txs to keep, default to be 50
                            keep_weeks_of_recent_txs            INT, -- number of weeks of recent txs to keep, default to be 1, business logic should be buffer 1 more week
                            expiry_at_epoch                     BIGINT, -- the epoch UTC seconds when the chain is expired

                            CONSTRAINT chain_info_pkey PRIMARY KEY (chain_id),
                            CONSTRAINT chain_info_unique_chain_name UNIQUE ("name") -- chain name must be unique
);

CREATE TABLE account (
                         chain_id        TEXT    NOT NULL,   -- also used as partition key
                         bech32_address  TEXT    NOT NULL,   -- normalized: lowercase

    -- inc by one per record inserted to `ref_account_to_recent_tx`,
    -- upon reaching specific number, prune the oldest ref and reset
                         continous_insert_ref_cur_tx_counter SMALLINT    NOT NULL DEFAULT 0,

    -- contracts which account has balance of
                         balance_on_erc20_contracts  TEXT[], -- normalized: lowercase, for both ERC-20 and CW-20
                         balance_on_nft_contracts    TEXT[], -- normalized: lowercase

                         CONSTRAINT account_pkey PRIMARY KEY (chain_id, bech32_address),
                         CONSTRAINT account_to_chain_info_fkey FOREIGN KEY (chain_id) REFERENCES chain_info(chain_id)
) PARTITION BY LIST(chain_id);
-- index for lookup account by bech32 address, multi-chain
CREATE INDEX account_b32_addr_index ON account (bech32_address);
-- trigger function for distinct the ERC-20 and CW-20 and NFT referenced contracts before insert or update account record
CREATE OR REPLACE FUNCTION func_trigger_00100_before_insert_or_update_account() RETURNS TRIGGER AS $$
BEGIN
    -- distinct the ERC-20 and CW-20 referenced contracts
    IF NEW.balance_on_erc20_contracts IS NOT NULL THEN
        NEW.balance_on_erc20_contracts := ARRAY(
            SELECT DISTINCT unnest(NEW.balance_on_erc20_contracts)
        );
END IF;

    -- distinct the NFT referenced contracts
    IF NEW.balance_on_nft_contracts IS NOT NULL THEN
        NEW.balance_on_nft_contracts := ARRAY(
            SELECT DISTINCT unnest(NEW.balance_on_nft_contracts)
        );
END IF;
RETURN NEW;
END;$$ LANGUAGE plpgsql;
CREATE TRIGGER trigger_00100_before_insert_or_update_account
    BEFORE INSERT OR UPDATE ON account
                         FOR EACH ROW EXECUTE FUNCTION func_trigger_00100_before_insert_or_update_account();

-- table recent_account_transaction
CREATE TABLE recent_account_transaction (
    -- main columns
                                            chain_id            TEXT        NOT NULL,   -- also used as partition key
                                            height              BIGINT      NOT NULL,
                                            hash                TEXT        NOT NULL,
                                            ref_count           SMALLINT    NOT NULL DEFAULT 0, -- number of references to this tx when reduced to zero, delete the tx.

    -- view-only columns
                                            epoch               BIGINT      NOT NULL, -- epoch UTC seconds
                                            message_types       TEXT[]      NOT NULL, -- proto message types of inner messages
                                            "action"            TEXT, -- action, probably available on evm/wasm txs. Generic values are "create", "transfer", "call:0x..."
                                            "value"             TEXT[], -- value of the transaction, eg: transfer amount, delegation amount, etc

                                            CONSTRAINT recent_account_transaction_pkey PRIMARY KEY (chain_id, height, hash)
) PARTITION BY LIST(chain_id);
-- trigger function for put recent_account_transaction into reduced_ref_count_recent_account_transaction
-- so if there is no reference to the tx, it will be pruned immediately.
CREATE OR REPLACE FUNCTION func_trigger_00100_after_insert_recent_account_transaction() RETURNS TRIGGER AS $$
BEGIN
INSERT INTO reduced_ref_count_recent_account_transaction(chain_id, height, hash)
VALUES (NEW.chain_id, NEW.height, NEW.hash)
    ON CONFLICT DO NOTHING;

RETURN NULL; -- result is ignored since this is an AFTER trigger
END;$$ LANGUAGE plpgsql;
CREATE TRIGGER trigger_00100_after_insert_recent_account_transaction
    AFTER INSERT ON recent_account_transaction
    FOR EACH ROW EXECUTE FUNCTION func_trigger_00100_after_insert_recent_account_transaction();

-- table reduced_ref_count_recent_account_transaction
-- A table with short-live records, used to cache records which reduced ref_count, then to prune corresponding record.
CREATE TABLE reduced_ref_count_recent_account_transaction (
                                                              chain_id            TEXT        NOT NULL,
                                                              height              BIGINT      NOT NULL,
                                                              hash                TEXT        NOT NULL,

                                                              CONSTRAINT reduced_ref_count_recent_account_transaction_pkey PRIMARY KEY (chain_id, height, hash)
);

-- table ref_account_to_recent_tx
CREATE TABLE ref_account_to_recent_tx (
                                          chain_id        TEXT    NOT NULL,   -- also used as partition key
                                          bech32_address  TEXT    NOT NULL,   -- normalized: lowercase
                                          height          BIGINT  NOT NULL,
                                          hash            TEXT    NOT NULL,

                                          signer          BOOLEAN, -- true if the address is one of the signers of the tx. `false` is not guaranteed to be the signer.
                                          erc20           BOOLEAN, -- true if the tx is erc20/cw20 tx
                                          nft             BOOLEAN, -- true if the tx is nft tx

                                          CONSTRAINT ref_account_to_recent_tx_pkey PRIMARY KEY (chain_id, bech32_address, height, hash),
                                          CONSTRAINT ref_recent_acc_tx_to_account_fkey FOREIGN KEY (chain_id, bech32_address)
                                              REFERENCES account(chain_id, bech32_address),
                                          CONSTRAINT ref_recent_acc_tx_to_recent_tx_fkey FOREIGN KEY (chain_id, height, hash)
                                              REFERENCES recent_account_transaction(chain_id, height, hash)
) PARTITION BY LIST(chain_id);
-- index for lookup recent tx by account, as well as for pruning
CREATE INDEX ref_account_to_recent_tx_by_account_index ON ref_account_to_recent_tx(chain_id, bech32_address);
-- trigger function for updating reference to tables account and recent_account_transaction after insert ref_account_to_recent_tx record
CREATE OR REPLACE FUNCTION func_trigger_00100_after_insert_ref_account_to_recent_tx() RETURNS TRIGGER AS $$
BEGIN
    -- increase reference count to account
UPDATE account SET continous_insert_ref_cur_tx_counter = continous_insert_ref_cur_tx_counter + 1
WHERE chain_id = NEW.chain_id AND bech32_address = NEW.bech32_address;

-- increase reference count to recent_account_transaction
UPDATE recent_account_transaction SET ref_count = ref_count + 1
WHERE chain_id = NEW.chain_id AND height = NEW.height AND hash = NEW.hash;

RETURN NULL; -- result is ignored since this is an AFTER trigger
END;$$ LANGUAGE plpgsql;
CREATE TRIGGER trigger_00100_after_insert_ref_account_to_recent_tx
    AFTER INSERT ON ref_account_to_recent_tx
    FOR EACH ROW EXECUTE FUNCTION func_trigger_00100_after_insert_ref_account_to_recent_tx();
-- trigger function for pruning recent_account_transaction after continous_insert_ref_cur_tx_counter reaches a specific number
CREATE OR REPLACE FUNCTION func_trigger_00200_after_insert_ref_account_to_recent_tx() RETURNS TRIGGER AS $$
DECLARE
later_continous_insert_ref_cur_tx_counter SMALLINT;
    -- prune the oldest ref_account_to_recent_tx record after X continuous insert,
    -- keep most recent X records for each type,
    -- for the corresponding account
    pruning_after_X_continous_insert CONSTANT INTEGER := 10;
    pruning_keep_recent_min_default CONSTANT INTEGER := 50;
    pruning_keep_recent INTEGER;
BEGIN
    -- get the pruning_keep_recent value from chain_info
SELECT GREATEST(COALESCE(ci.keep_recent_account_tx_count, pruning_keep_recent_min_default), pruning_keep_recent_min_default)
INTO pruning_keep_recent
FROM chain_info ci WHERE ci.chain_id = NEW.chain_id;

-- check if the counter reaches a specific number
SELECT acc.continous_insert_ref_cur_tx_counter INTO later_continous_insert_ref_cur_tx_counter
FROM account acc WHERE acc.chain_id = NEW.chain_id AND acc.bech32_address = NEW.bech32_address;
IF later_continous_insert_ref_cur_tx_counter >= pruning_after_X_continous_insert THEN
        -- prune the oldest ref_account_to_recent_tx record
DELETE FROM ref_account_to_recent_tx
WHERE chain_id = NEW.chain_id AND bech32_address = NEW.bech32_address
  AND height NOT IN (
    SELECT DISTINCT(hs.height) FROM (
                                        -- keep most recent normal txs
                                        (
                                            SELECT height FROM ref_account_to_recent_tx
                                            WHERE chain_id = NEW.chain_id AND bech32_address = NEW.bech32_address
                                            ORDER BY height DESC
                                                LIMIT pruning_keep_recent
                                        )
                                        -- keep most recent sent txs
                                        UNION
                                        (
                                            SELECT height FROM ref_account_to_recent_tx
                                            WHERE chain_id = NEW.chain_id AND bech32_address = NEW.bech32_address AND signer IS TRUE
                                            ORDER BY height DESC
                                                LIMIT pruning_keep_recent
                                        )
                                        -- keep most recent erc20/cw20 txs
                                        UNION
                                        (
                                            SELECT height FROM ref_account_to_recent_tx
                                            WHERE chain_id = NEW.chain_id AND bech32_address = NEW.bech32_address AND erc20 IS TRUE
                                            ORDER BY height DESC
                                                LIMIT pruning_keep_recent
                                        )
                                        -- keep most recent nft txs
                                        UNION
                                        (
                                            SELECT height FROM ref_account_to_recent_tx
                                            WHERE chain_id = NEW.chain_id AND bech32_address = NEW.bech32_address AND nft IS TRUE
                                            ORDER BY height DESC
                                                LIMIT pruning_keep_recent
                                        )
                                    ) hs
);

-- reset the counter
UPDATE account SET continous_insert_ref_cur_tx_counter = 0
WHERE chain_id = NEW.chain_id AND bech32_address = NEW.bech32_address;
END IF;

RETURN NULL; -- result is ignored since this is an AFTER trigger
END;$$ LANGUAGE plpgsql;
CREATE TRIGGER trigger_00200_after_insert_ref_account_to_recent_tx
    AFTER INSERT ON ref_account_to_recent_tx
    FOR EACH ROW EXECUTE FUNCTION func_trigger_00200_after_insert_ref_account_to_recent_tx();
-- trigger function for reducing reference on recent_account_transaction after delete ref_account_to_recent_tx record
CREATE OR REPLACE FUNCTION func_trigger_00300_after_delete_ref_account_to_recent_tx() RETURNS TRIGGER AS $$
BEGIN
    -- reduce reference count
UPDATE recent_account_transaction SET ref_count = ref_count - 1
WHERE chain_id = OLD.chain_id AND height = OLD.height AND hash = OLD.hash;

INSERT INTO reduced_ref_count_recent_account_transaction(chain_id, height, hash)
VALUES (OLD.chain_id, OLD.height, OLD.hash)
    ON CONFLICT DO NOTHING;

RETURN NULL; -- result is ignored since this is an AFTER trigger
END;$$ LANGUAGE plpgsql;
CREATE TRIGGER trigger_00300_after_delete_ref_account_to_recent_tx
    AFTER DELETE ON ref_account_to_recent_tx
    FOR EACH ROW EXECUTE FUNCTION func_trigger_00300_after_delete_ref_account_to_recent_tx();
-- procedure for pruning recent_account_transaction after update ref count to zero
CREATE OR REPLACE PROCEDURE func_cleanup_zero_ref_count_recent_account_transaction() AS $$
DECLARE
reduced RECORD;
    current_ref_count SMALLINT;
BEGIN
FOR reduced IN (SELECT rr.chain_id, rr.height, rr.hash FROM reduced_ref_count_recent_account_transaction rr)
    LOOP
SELECT ref_count INTO current_ref_count FROM recent_account_transaction
WHERE chain_id = reduced.chain_id AND height = reduced.height AND hash = reduced.hash;
IF current_ref_count < 1 THEN
DELETE FROM recent_account_transaction
WHERE chain_id = reduced.chain_id AND height = reduced.height AND hash = reduced.hash;
END IF;
DELETE FROM reduced_ref_count_recent_account_transaction
WHERE chain_id = reduced.chain_id AND height = reduced.height AND hash = reduced.hash;
END LOOP;
END;$$ LANGUAGE plpgsql;

-- table transaction
-- Page: search multi-chain transactions, search single-chain, showing blocks & transactions list
CREATE TABLE transaction (
    -- pk fields
                             chain_id            TEXT    NOT NULL,
                             height              BIGINT  NOT NULL,
                             hash                TEXT    NOT NULL, -- normalized: Cosmos: uppercase without 0x, Ethereum: lowercase with 0x
                             partition_id        TEXT    NOT NULL, -- `${epoch week}_${chain_id}` (epoch week = FLOOR(epoch UTC seconds / (3600 sec x 24 hours x 7 days)))

    -- other fields
                             epoch               BIGINT  NOT NULL, -- epoch UTC seconds
                             message_types       TEXT[]  NOT NULL, -- proto message types of inner messages
                             tx_type             TEXT    NOT NULL, -- tx type, eg: cosmos or evm or wasm
                             "action"            TEXT, -- action, probably available on evm/wasm txs. Generic values are "create", "transfer", "call:0x..."
                             "value"             TEXT[], -- value of the transaction, eg: transfer amount, delegation amount, etc

                             CONSTRAINT transaction_pkey PRIMARY KEY (chain_id, height, hash, partition_id),
                             CONSTRAINT transaction_to_chain_info_fkey FOREIGN KEY (chain_id) REFERENCES chain_info(chain_id)
) PARTITION BY LIST(partition_id);
-- index for lookup transaction by hash, multi-chain & single-chain
CREATE INDEX transaction_hash_index ON transaction(hash);

-- table ibc_transaction
-- Data on this table should be stored permanently, as it is used for mapping IBC transactions.
CREATE TABLE ibc_transaction
(
    -- pk fields
    chain_id                TEXT    NOT NULL,
    height                  BIGINT  NOT NULL,
    hash                    TEXT    NOT NULL,
    -- other fields
    "type"                  TEXT    NOT NULL, -- "TRF", "RECV", "ACK", "TO"
    sequence_no             TEXT    NOT NULL,
    port                    TEXT    NOT NULL,
    channel                 TEXT    NOT NULL,
    counter_party_port      TEXT    NOT NULL,
    counter_party_channel   TEXT    NOT NULL,
    incoming                BOOLEAN,          -- true if the transaction is incoming (MsgRecvPacket), false if outgoing
    CONSTRAINT ibc_transaction_pkey PRIMARY KEY (chain_id, height, hash)
) PARTITION BY LIST(chain_id);
CREATE INDEX ibctx_same_sequence_index ON ibc_transaction(chain_id, sequence_no, port, channel, incoming);

-- table failed_block
-- For storing failed to index - blocks
CREATE TABLE failed_block (
    -- pk fields
                              chain_id    TEXT    NOT NULL, -- also used as partition key
                              height      BIGINT  NOT NULL,

    -- logic fields
                              retry_count         SMALLINT    NOT NULL DEFAULT 0, -- number of retry to index the block
                              last_retry_epoch    BIGINT      NOT NULL DEFAULT 0, -- last retry UTC seconds

    -- information fields
                              error_messages       TEXT[]      NOT NULL DEFAULT '{}', -- error messages when failed to index the block

                              CONSTRAINT failed_block_pkey PRIMARY KEY (chain_id, height)
) PARTITION BY LIST(chain_id);

-- table partition_table_info
CREATE TABLE partition_table_info (
                                      partition_table_name    TEXT    NOT NULL,
                                      large_table_name        TEXT    NOT NULL,

    -- for information only
                                      partition_key           TEXT    NOT NULL, -- string representation of partition keys

    -- partition key parts, is part 1 only if single partition key, part 2 is optional when multi-key combined
                                      partition_key_part_1    TEXT    NOT NULL,
                                      partition_key_part_2    TEXT,
                                      CONSTRAINT partition_table_info_pkey PRIMARY KEY (partition_table_name)
);
CREATE INDEX pti_table_and_key1_index ON partition_table_info(large_table_name, partition_key_part_1);

-- Helper methods

-- function get_indexing_fallbehind_chains
-- Used to get the list of chains which indexed block height is behind the current time by more than a specific threshold
CREATE OR REPLACE FUNCTION get_indexing_fallbehind_chains(threshold_seconds BIGINT) RETURNS TABLE(chain_id TEXT, height BIGINT, epoch BIGINT, epoch_diff BIGINT) AS $$
DECLARE
epoch_utc_now BIGINT;
BEGIN
SELECT FLOOR(EXTRACT(epoch FROM NOW() AT TIME ZONE 'utc' AT TIME ZONE 'utc')::BIGINT)::BIGINT INTO epoch_utc_now;
RETURN QUERY SELECT ci.chain_id, ci.height, ci.epoch, ci.epoch_diff FROM (
		SELECT
			i.chain_id,
			i.latest_indexed_block AS height,
			i.increased_latest_indexed_block_at AS epoch,
			epoch_utc_now - i.increased_latest_indexed_block_at AS epoch_diff
		FROM chain_info i
		WHERE i.postponed IS NOT TRUE AND (i.expiry_at_epoch IS NULL OR i.expiry_at_epoch > epoch_utc_now)
	) ci
	WHERE ci.epoch_diff > threshold_seconds
	ORDER BY ci.epoch_diff DESC;
END;$$ LANGUAGE plpgsql;

-- Used to get the list of partitioned tables of the large table "transaction" which should be pruned, based on configuration and current epoch
CREATE OR REPLACE FUNCTION get_partitioned_transaction_tables_to_prune(epoch_utc_now BIGINT) RETURNS TABLE(partition_table_name TEXT) AS $$
DECLARE
epoch_input_vs_server_diff BIGINT;
	current_epoch_week INT;
BEGIN
	-- validate input
SELECT ABS(FLOOR(EXTRACT(epoch FROM NOW() AT TIME ZONE 'utc' AT TIME ZONE 'utc')::BIGINT)::BIGINT - epoch_utc_now) INTO epoch_input_vs_server_diff;
IF epoch_input_vs_server_diff > 86400 THEN
		RAISE EXCEPTION 'provided epoch UTC vs server epoch UTC has big tolerant %', epoch_input_vs_server_diff;
END IF;

SELECT FLOOR(epoch_utc_now / (86400 * 7))::INT INTO current_epoch_week;

--
RETURN QUERY SELECT a.partition_table_name FROM (
		SELECT ci.chain_id, GREATEST(1, COALESCE(ci.keep_weeks_of_recent_txs, 1)) + 1 AS keep_weeks_of_recent_txs_with_buffer, sub_pti.*
		FROM chain_info ci
		LEFT JOIN (
			SELECT pti.partition_table_name, pti.partition_key_part_1, pti.partition_key_part_2
			FROM partition_table_info pti
			WHERE pti.large_table_name = 'transaction'
		) sub_pti
		ON ci.chain_id = sub_pti.partition_key_part_2
	) a WHERE current_epoch_week - a.partition_key_part_1::INT >= a.keep_weeks_of_recent_txs_with_buffer;
END;$$ LANGUAGE plpgsql;
