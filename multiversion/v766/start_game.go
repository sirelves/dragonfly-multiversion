/*
 _____               _____
|_   _| __ _   _  __|_   _|__ _ __ ___  _ __ ___
  | || '__| | | |/ _ \| |/ _ \ '_ ` _ \| '_ ` _ \
  | || |  | |_| |  __/| |  __/ | | | | | | | | | |
  |_||_|   \__,_|\___||_|\___|_| |_| |_|_| |_| |_|

 _____ __  __       ____                               __ _
|_   _|  \/  |     |  _ \ _ __ __ _  __ _  ___  _ __  / _| |_   _
  | | | |\/| |_____| | | | '__/ _` |/ _` |/ _ \| '_ \| |_| | | | |
  | | | |  | |_____| |_| | | | (_| | (_| | (_) | | | |  _| | |_| |
  |_| |_|  |_|     |____/|_|  \__,_|\__, |\___/|_| |_|_| |_|\__, |
                                    |___/                   |___/

@author TrueTemm
@link   https://github.com/TrueTemm
TM-Dragonfly Project
*/

package v766

import (
	"github.com/df-mc/dragonfly/multiversion/v786"
	"github.com/sandertv/gophertunnel/minecraft/nbt"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

const (
	SpawnBiomeTypeDefault = iota
	SpawnBiomeTypeUserDefined
)

const (
	ChatRestrictionLevelNone     = 0
	ChatRestrictionLevelDropped  = 1
	ChatRestrictionLevelDisabled = 2
)

const (
	EditorWorldTypeNotEditor = iota
	EditorWorldTypeProject
	EditorWorldTypeTestLevel
)

type ItemEntry struct {
	Name           string
	RuntimeID      int16
	ComponentBased bool
}

func (x *ItemEntry) Marshal(r protocol.IO) {
	r.String(&x.Name)
	r.Int16(&x.RuntimeID)
	if p := v786.ProtoOf(r); p == 0 || p >= 419 { // added in 419
		r.Bool(&x.ComponentBased)
	}
}

type StartGame struct {
	v786.StartGame

	Items []ItemEntry

	// LegacyBlockPalette is the network NBT list of every block state, sent
	// instead of Blocks to protocols before 419.
	LegacyBlockPalette []byte
}

func (*StartGame) ID() uint32 {
	return v786.IDStartGame
}

func (pk *StartGame) Marshal(io protocol.IO) {
	io.Varint64(&pk.EntityUniqueID)
	io.Varuint64(&pk.EntityRuntimeID)
	io.Varint32(&pk.PlayerGameMode)
	io.Vec3(&pk.PlayerPosition)
	io.Float32(&pk.Pitch)
	io.Float32(&pk.Yaw)
	if p := v786.ProtoOf(io); p != 0 && p < 503 { // a varint32 before 503
		seed := int32(pk.WorldSeed)
		io.Varint32(&seed)
		pk.WorldSeed = int64(seed)
	} else {
		io.Int64(&pk.WorldSeed)
	}
	io.Int16(&pk.SpawnBiomeType)
	io.String(&pk.UserDefinedBiomeName)
	io.Varint32(&pk.Dimension)
	io.Varint32(&pk.Generator)
	io.Varint32(&pk.WorldGameMode)
	if p := v786.ProtoOf(io); p == 0 || p >= 671 { // added in 671
		io.Bool(&pk.Hardcore)
	}
	io.Varint32(&pk.Difficulty)
	v786.UBlockPos786(io, &pk.WorldSpawn)
	io.Bool(&pk.AchievementsDisabled)
	if p := v786.ProtoOf(io); p != 0 && p < 534 {
		// no editor flag before 534
	} else if p != 0 && p < 618 {
		editor := pk.EditorWorldType != 0
		io.Bool(&editor)
		if editor {
			pk.EditorWorldType = 1
		} else {
			pk.EditorWorldType = 0
		}
	} else {
		io.Varint32(&pk.EditorWorldType)
	}
	if p := v786.ProtoOf(io); p == 0 || p >= 582 { // added in 582
		io.Bool(&pk.CreatedInEditor)
		io.Bool(&pk.ExportedFromEditor)
	}
	io.Varint32(&pk.DayCycleLockTime)
	io.Varint32(&pk.EducationEditionOffer)
	io.Bool(&pk.EducationFeaturesEnabled)
	io.String(&pk.EducationProductID)
	io.Float32(&pk.RainLevel)
	io.Float32(&pk.LightningLevel)
	io.Bool(&pk.ConfirmedPlatformLockedContent)
	io.Bool(&pk.MultiPlayerGame)
	io.Bool(&pk.LANBroadcastEnabled)
	io.Varint32(&pk.XBLBroadcastMode)
	io.Varint32(&pk.PlatformBroadcastMode)
	io.Bool(&pk.CommandsEnabled)
	io.Bool(&pk.TexturePackRequired)
	protocol.FuncSlice(io, &pk.GameRules, io.GameRule)
	if p := v786.ProtoOf(io); p == 0 || p >= 419 { // added in 419
		protocol.SliceUint32Length(io, &pk.Experiments)
		io.Bool(&pk.ExperimentsPreviouslyToggled)
	}
	io.Bool(&pk.BonusChestEnabled)
	io.Bool(&pk.StartWithMapEnabled)
	io.Varint32(&pk.PlayerPermissions)
	io.Int32(&pk.ServerChunkTickRadius)
	io.Bool(&pk.HasLockedBehaviourPack)
	io.Bool(&pk.HasLockedTexturePack)
	io.Bool(&pk.FromLockedWorldTemplate)
	io.Bool(&pk.MSAGamerTagsOnly)
	io.Bool(&pk.FromWorldTemplate)
	io.Bool(&pk.WorldTemplateSettingsLocked)
	io.Bool(&pk.OnlySpawnV1Villagers)
	if p := v786.ProtoOf(io); p == 0 || p >= 544 { // added in 544
		io.Bool(&pk.PersonaDisabled)
		io.Bool(&pk.CustomSkinsDisabled)
	}
	if p := v786.ProtoOf(io); p == 0 || p >= 567 { // added in 567
		io.Bool(&pk.EmoteChatMuted)
	}
	io.String(&pk.BaseGameVersion)
	io.Int32(&pk.LimitedWorldWidth)
	io.Int32(&pk.LimitedWorldDepth)
	io.Bool(&pk.NewNether)
	if p := v786.ProtoOf(io); p == 0 || p >= 465 { // added in 465
		protocol.Single(io, &pk.EducationSharedResourceURI)
	}
	protocol.OptionalFunc(io, &pk.ForceExperimentalGameplay, io.Bool)
	if p := v786.ProtoOf(io); p == 0 || p >= 544 { // added in 544
		io.Uint8(&pk.ChatRestrictionLevel)
		io.Bool(&pk.DisablePlayerInteractions)
	}
	if p := v786.ProtoOf(io); p == 0 || p >= 685 { // added in 685
		io.String(&pk.ServerID)
		io.String(&pk.WorldID)
		io.String(&pk.ScenarioID)
	}
	io.String(&pk.LevelID)
	io.String(&pk.WorldName)
	io.String(&pk.TemplateContentIdentity)
	io.Bool(&pk.Trial)
	if p := v786.ProtoOf(io); p != 0 && p < 419 { // a server authoritative movement bool before 419
		server := pk.PlayerMovementSettings.MovementType != 0
		io.Bool(&server)
		if server && pk.PlayerMovementSettings.MovementType == 0 {
			pk.PlayerMovementSettings.MovementType = 1
		}
	} else if p != 0 && p < 428 { // only the movement mode before 428
		mode := uint32(pk.PlayerMovementSettings.MovementType)
		io.Varuint32(&mode)
		pk.PlayerMovementSettings.MovementType = int32(mode)
	} else {
		v786.PlayerMoveSettings786(io, &pk.PlayerMovementSettings)
	}
	io.Int64(&pk.Time)
	io.Varint32(&pk.EnchantmentSeed)
	if p := v786.ProtoOf(io); p != 0 && p < 419 {
		// The whole block palette as an NBT list, which also fixes the runtime IDs.
		if _, reading := io.(v786.Reader786); reading {
			var list []any
			io.NBTList(&list, nbt.NetworkLittleEndian)
			pk.LegacyBlockPalette, _ = nbt.MarshalEncoding(list, nbt.NetworkLittleEndian)
		} else {
			io.Bytes(&pk.LegacyBlockPalette)
		}
	} else {
		protocol.Slice(io, &pk.Blocks)
	}
	protocol.Slice(io, &pk.Items)
	io.String(&pk.MultiPlayerCorrelationID)
	io.Bool(&pk.ServerAuthoritativeInventory)
	if p := v786.ProtoOf(io); p == 0 || p >= 440 { // added in 440
		io.String(&pk.GameVersion)
	}
	if p := v786.ProtoOf(io); p == 0 || p >= 527 { // added in 527
		io.NBT(&pk.PropertyData, nbt.NetworkLittleEndian)
	}
	if p := v786.ProtoOf(io); p == 0 || p >= 475 { // added in 475
		io.Uint64(&pk.ServerBlockStateChecksum)
	}
	if p := v786.ProtoOf(io); p == 0 || p >= 527 { // added in 527
		io.UUID(&pk.WorldTemplateID)
	}
	if p := v786.ProtoOf(io); p == 0 || p >= 544 { // added in 544
		io.Bool(&pk.ClientSideGeneration)
	}
	if p := v786.ProtoOf(io); p == 0 || p >= 582 { // added in 582
		io.Bool(&pk.UseBlockNetworkIDHashes)
	}
	if p := v786.ProtoOf(io); p == 0 || p >= 589 { // added in 589
		io.Bool(&pk.ServerAuthoritativeSound)
	}
}
