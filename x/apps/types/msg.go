package types

import (
	"github.com/pokt-network/posmint/crypto"
	sdk "github.com/pokt-network/posmint/types"
)

// ensure Msg interface compliance at compile time
var (
	_ sdk.Msg = &MsgAppStake{}
	_ sdk.Msg = &MsgBeginAppUnstake{}
	_ sdk.Msg = &MsgAppUnjail{}
)

//----------------------------------------------------------------------------------------------------------------------
// MsgAppStake - struct for staking transactions
type MsgAppStake struct {
	Address sdk.Address      `json:"application_address" yaml:"application_address"`
	PubKey  crypto.PublicKey `json:"pubkey" yaml:"pubkey"`
	Chains  []string         `json:"chains" yaml:"chains"`
	Value   sdk.Int          `json:"value" yaml:"value"`
}

// Return address(es) that must sign over msg.GetSignBytes()
func (msg MsgAppStake) GetSigners() []sdk.Address {
	addrs := []sdk.Address{sdk.Address(msg.Address)}
	return addrs
}

// GetSignBytes returns the message bytes to sign over.
func (msg MsgAppStake) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// quick validity check
func (msg MsgAppStake) ValidateBasic() sdk.Error {
	if msg.Address.Empty() {
		return ErrNilApplicationAddr(DefaultCodespace)
	}
	if msg.Value.LTE(sdk.ZeroInt()) {
		return ErrBadStakeAmount(DefaultCodespace)
	}
	if len(msg.Chains) == 0 {
		return ErrNoChains(DefaultCodespace)
	}
	for _, chain := range msg.Chains {
		if len(chain) == 0 {
			return ErrNoChains(DefaultCodespace)
		}
	}
	return nil
}

//nolint
func (msg MsgAppStake) Route() string { return RouterKey }
func (msg MsgAppStake) Type() string  { return "stake_application" }

//----------------------------------------------------------------------------------------------------------------------
// MsgBeginAppUnstake - struct for unstaking transaciton
type MsgBeginAppUnstake struct {
	Address sdk.Address `json:"application_address" yaml:"application_address"`
}

func (msg MsgBeginAppUnstake) GetSigners() []sdk.Address {
	return []sdk.Address{sdk.Address(msg.Address)}
}

func (msg MsgBeginAppUnstake) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg MsgBeginAppUnstake) ValidateBasic() sdk.Error {
	if msg.Address.Empty() {
		return ErrNilApplicationAddr(DefaultCodespace)
	}
	return nil
}

//nolint
func (msg MsgBeginAppUnstake) Route() string { return RouterKey }
func (msg MsgBeginAppUnstake) Type() string  { return "begin_unstaking_application" }

//----------------------------------------------------------------------------------------------------------------------
// MsgAppUnjail - struct for unjailing jailed application
type MsgAppUnjail struct {
	AppAddr sdk.Address `json:"address" yaml:"address"` // address of the application operator
}

//nolint
func (msg MsgAppUnjail) Route() string { return RouterKey }
func (msg MsgAppUnjail) Type() string  { return "unjail" }
func (msg MsgAppUnjail) GetSigners() []sdk.Address {
	return []sdk.Address{sdk.Address(msg.AppAddr)}
}

func (msg MsgAppUnjail) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg MsgAppUnjail) ValidateBasic() sdk.Error {
	if msg.AppAddr.Empty() {
		return ErrBadApplicationAddr(DefaultCodespace)
	}
	return nil
}
