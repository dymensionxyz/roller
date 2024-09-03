CREATE TABLE "account" (
    "chain_id" TEXT NOT NULL,
    "bech32_address" TEXT NOT NULL,
    "continous_insert_ref_cur_tx_counter" SMALLINT NOT NULL DEFAULT 0,
    "balance_on_erc20_contracts" TEXT[],
    "balance_on_nft_contracts" TEXT[],

    CONSTRAINT "account_pkey" PRIMARY KEY ("chain_id","bech32_address")
);

CREATE TABLE "chain_info" (
    "chain_id" TEXT NOT NULL,
    "name" TEXT NOT NULL,
    "chain_type" TEXT NOT NULL,
    "bech32" JSONB NOT NULL,
    "denoms" JSONB NOT NULL,
    "be_json_rpc_urls" TEXT[],
    "latest_indexed_block" BIGINT NOT NULL DEFAULT 0,
    "increased_latest_indexed_block_at" BIGINT NOT NULL DEFAULT 0,
    "postponed" BOOLEAN,
    "keep_recent_account_tx_count" INTEGER,
    "expiry_at_epoch" BIGINT,
    "keep_weeks_of_recent_txs" INTEGER,

    CONSTRAINT "chain_info_pkey" PRIMARY KEY ("chain_id")
);

CREATE TABLE "ref_account_to_recent_tx" (
    "chain_id" TEXT NOT NULL,
    "bech32_address" TEXT NOT NULL,
    "height" BIGINT NOT NULL,
    "hash" TEXT NOT NULL,
    "signer" BOOLEAN,
    "erc20" BOOLEAN,
    "nft" BOOLEAN,

    CONSTRAINT "ref_account_to_recent_tx_pkey" PRIMARY KEY ("chain_id","bech32_address","height","hash")
);

CREATE TABLE "transaction" (
    "chain_id" TEXT NOT NULL,
    "height" BIGINT NOT NULL,
    "hash" TEXT NOT NULL,
    "partition_id" TEXT NOT NULL,
    "epoch" BIGINT NOT NULL,
    "message_types" TEXT[],
    "tx_type" TEXT NOT NULL,
    "action" TEXT,
    "value" TEXT[],

    CONSTRAINT "transaction_pkey" PRIMARY KEY ("chain_id","height","hash","partition_id")
);

CREATE TABLE "failed_block" (
    "chain_id" TEXT NOT NULL,
    "height" BIGINT NOT NULL,
    "retry_count" SMALLINT NOT NULL DEFAULT 0,
    "last_retry_epoch" BIGINT NOT NULL DEFAULT 0,
    "error_messages" TEXT[] DEFAULT ARRAY[]::TEXT[],

    CONSTRAINT "failed_block_pkey" PRIMARY KEY ("chain_id","height")
);

CREATE TABLE "recent_account_transaction" (
    "chain_id" TEXT NOT NULL,
    "height" BIGINT NOT NULL,
    "hash" TEXT NOT NULL,
    "ref_count" SMALLINT NOT NULL DEFAULT 0,
    "epoch" BIGINT NOT NULL,
    "message_types" TEXT[],
    "action" TEXT,
    "value" TEXT[],

    CONSTRAINT "recent_account_transaction_pkey" PRIMARY KEY ("chain_id","height","hash")
);

CREATE TABLE "reduced_ref_count_recent_account_transaction" (
    "chain_id" TEXT NOT NULL,
    "height" BIGINT NOT NULL,
    "hash" TEXT NOT NULL,

    CONSTRAINT "reduced_ref_count_recent_account_transaction_pkey" PRIMARY KEY ("chain_id","height","hash")
);

CREATE TABLE "ibc_transaction" (
    "chain_id" TEXT NOT NULL,
    "height" BIGINT NOT NULL,
    "hash" TEXT NOT NULL,
    "type" TEXT NOT NULL,
    "sequence_no" TEXT NOT NULL,
    "port" TEXT NOT NULL,
    "channel" TEXT NOT NULL,
    "counter_party_port" TEXT NOT NULL,
    "counter_party_channel" TEXT NOT NULL,
    "incoming" BOOLEAN,

    CONSTRAINT "ibc_transaction_pkey" PRIMARY KEY ("chain_id","height","hash")
);

CREATE TABLE "partition_table_info" (
    "partition_table_name" TEXT NOT NULL,
    "large_table_name" TEXT NOT NULL,
    "partition_key" TEXT NOT NULL,
    "partition_key_part_1" TEXT NOT NULL,
    "partition_key_part_2" TEXT,

    CONSTRAINT "partition_table_info_pkey" PRIMARY KEY ("partition_table_name")
);

CREATE INDEX "account_b32_addr_index" ON "account"("bech32_address");

CREATE UNIQUE INDEX "chain_info_unique_chain_name" ON "chain_info"("name");

CREATE INDEX "ref_account_to_recent_tx_by_account_index" ON "ref_account_to_recent_tx"("chain_id", "bech32_address");

CREATE INDEX "transaction_hash_index" ON "transaction"("hash");

CREATE INDEX "ibctx_same_sequence_index" ON "ibc_transaction"("chain_id", "sequence_no", "port", "channel", "incoming");

CREATE INDEX "pti_table_and_key1_index" ON "partition_table_info"("large_table_name", "partition_key_part_1");

ALTER TABLE "account" ADD CONSTRAINT "account_to_chain_info_fkey" FOREIGN KEY ("chain_id") REFERENCES "chain_info"("chain_id") ON DELETE NO ACTION ON UPDATE NO ACTION;

ALTER TABLE "ref_account_to_recent_tx" ADD CONSTRAINT "ref_recent_acc_tx_to_account_fkey" FOREIGN KEY ("chain_id", "bech32_address") REFERENCES "account"("chain_id", "bech32_address") ON DELETE NO ACTION ON UPDATE NO ACTION;

ALTER TABLE "ref_account_to_recent_tx" ADD CONSTRAINT "ref_recent_acc_tx_to_recent_tx_fkey" FOREIGN KEY ("chain_id", "height", "hash") REFERENCES "recent_account_transaction"("chain_id", "height", "hash") ON DELETE NO ACTION ON UPDATE NO ACTION;

ALTER TABLE "transaction" ADD CONSTRAINT "transaction_to_chain_info_fkey" FOREIGN KEY ("chain_id") REFERENCES "chain_info"("chain_id") ON DELETE NO ACTION ON UPDATE NO ACTION;