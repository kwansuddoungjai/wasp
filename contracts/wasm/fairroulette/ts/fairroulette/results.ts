// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// (Re-)generated by schema tool
// >>>> DO NOT CHANGE THIS FILE! <<<<
// Change the json schema instead

import * as wasmlib from "wasmlib";
import * as sc from "./index";

export class ImmutableLastWinningNumberResults extends wasmlib.ScMapID {
    lastWinningNumber(): wasmlib.ScImmutableInt64 {
		return new wasmlib.ScImmutableInt64(this.mapID, sc.idxMap[sc.IdxResultLastWinningNumber]);
	}
}

export class MutableLastWinningNumberResults extends wasmlib.ScMapID {
    lastWinningNumber(): wasmlib.ScMutableInt64 {
		return new wasmlib.ScMutableInt64(this.mapID, sc.idxMap[sc.IdxResultLastWinningNumber]);
	}
}

export class ImmutableRoundNumberResults extends wasmlib.ScMapID {
    roundNumber(): wasmlib.ScImmutableInt64 {
		return new wasmlib.ScImmutableInt64(this.mapID, sc.idxMap[sc.IdxResultRoundNumber]);
	}
}

export class MutableRoundNumberResults extends wasmlib.ScMapID {
    roundNumber(): wasmlib.ScMutableInt64 {
		return new wasmlib.ScMutableInt64(this.mapID, sc.idxMap[sc.IdxResultRoundNumber]);
	}
}

export class ImmutableRoundStartedAtResults extends wasmlib.ScMapID {
    roundStartedAt(): wasmlib.ScImmutableInt32 {
		return new wasmlib.ScImmutableInt32(this.mapID, sc.idxMap[sc.IdxResultRoundStartedAt]);
	}
}

export class MutableRoundStartedAtResults extends wasmlib.ScMapID {
    roundStartedAt(): wasmlib.ScMutableInt32 {
		return new wasmlib.ScMutableInt32(this.mapID, sc.idxMap[sc.IdxResultRoundStartedAt]);
	}
}

export class ImmutableRoundStatusResults extends wasmlib.ScMapID {
    roundStatus(): wasmlib.ScImmutableInt16 {
		return new wasmlib.ScImmutableInt16(this.mapID, sc.idxMap[sc.IdxResultRoundStatus]);
	}
}

export class MutableRoundStatusResults extends wasmlib.ScMapID {
    roundStatus(): wasmlib.ScMutableInt16 {
		return new wasmlib.ScMutableInt16(this.mapID, sc.idxMap[sc.IdxResultRoundStatus]);
	}
}
