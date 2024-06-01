package cmd

import (
	"context"
	"fmt"
	units "github.com/docker/go-units"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/olekukonko/tablewriter"
	pb "github.com/spacemeshos/api/release/go/spacemesh/v1"
	"github.com/spacemeshos/go-spacemesh/common/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
	_ "net/http/pprof"
	"os"
	"strings"
)

var overviewCmd = &cobra.Command{
	Use:   "overview",
	Short: "print overview of smesher",
	Run: func(c *cobra.Command, args []string) {
		ctx := context.Background()
		if err := showOverviewStatus(ctx); err != nil {
			fmt.Println(err.Error())
		}
	},
}

func showOverviewStatus(ctx context.Context) error {
	privateURL := resolveToLocalhost(viper.GetString("api.grpc-private-listener"))
	publicURL := resolveToLocalhost(viper.GetString("api.grpc-public-listener"))

	overViewState, err := getOverViewState(ctx, publicURL, privateURL)
	if err != nil {
		return err
	}

	percent := fmt.Sprintf("%.2f %%", 100*(float64(overViewState.CompletedSize)/float64(overViewState.CommitmentSize)))
	commitStatus := fmt.Sprintf("%s / %s  %s", overViewState.CompletedSize.String(), overViewState.CommitmentSize.String(), percent)
	var syncStatus string
	if overViewState.IsSynced {
		syncStatus = fmt.Sprintf("Success Current(%d)GenesisEndLayer(%d)", overViewState.CurrentLayerId, overViewState.GenesisEndLayer)
	} else {
		syncStatus = fmt.Sprintf("Fail Current(%d) GenesisEndLayer(%d)", overViewState.CurrentLayerId, overViewState.GenesisEndLayer)
	}

	tbl := tablewriter.NewWriter(os.Stdout)
	tbl.Append([]string{"SmeshId", fmt.Sprintf("ID(0x%s) Addr(%s)", types.BytesToNodeID(overViewState.SmesherId).String(), types.GenerateAddress(overViewState.SmesherId).String())})
	tbl.Append([]string{"SmeshState", overViewState.State.String()})
	tbl.Append([]string{"CoinBase", overViewState.CoinBase})
	tbl.Append([]string{"GenesisId", overViewState.GenesisId.String()})
	tbl.Append([]string{"SyncStatus", syncStatus})
	tbl.Append([]string{"SmeshProgress", commitStatus})
	tbl.Render()
	return nil
}

func getOverViewState(ctx context.Context, publicURL, privateURL string) (*OverViewStatus, error) {
	publicConn, err := grpc.Dial(publicURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	privateConn, err := grpc.Dial(privateURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	smesherClient := pb.NewSmesherServiceClient(privateConn)
	status, err := smesherClient.PostSetupStatus(ctx, &empty.Empty{})
	if err != nil {
		return nil, err
	}

	postCfg, err := smesherClient.PostConfig(ctx, &empty.Empty{})
	if err != nil {
		return nil, err
	}

	smeshId, err := smesherClient.SmesherID(ctx, &emptypb.Empty{})
	if err != nil {
		return nil, err
	}

	coinBase, err := smesherClient.Coinbase(ctx, &empty.Empty{})
	if err != nil {
		return nil, err
	}

	nodeClient := pb.NewNodeServiceClient(publicConn)
	nodeStatus, err := nodeClient.Status(ctx, &pb.StatusRequest{})
	if err != nil {
		return nil, err
	}

	nodeInfo, err := nodeClient.NodeInfo(ctx, &empty.Empty{})
	if err != nil {
		return nil, err
	}

	meshClient := pb.NewMeshServiceClient(publicConn)
	genesisIDResp, err := meshClient.GenesisID(ctx, &pb.GenesisIDRequest{})
	if err != nil {
		return nil, err
	}

	commitmentSize := postCfg.LabelsPerUnit * uint64(postCfg.BitsPerLabel) * (uint64(status.GetStatus().GetOpts().NumUnits)) / 8
	completed := status.GetStatus().NumLabelsWritten * uint64(postCfg.BitsPerLabel) / 8

	v := types.Hash20{}
	copy(v[:], genesisIDResp.GenesisId)
	return &OverViewStatus{
		CompletedSize:   StorageSize(completed),
		CommitmentSize:  StorageSize(commitmentSize),
		State:           status.GetStatus().GetState(),
		CoinBase:        coinBase.AccountId.Address,
		SmesherId:       smeshId.PublicKey,
		GenesisId:       v,
		IsSynced:        nodeStatus.GetStatus().IsSynced,
		CurrentLayerId:  int(nodeStatus.GetStatus().TopLayer.Number),
		GenesisEndLayer: int(nodeInfo.GetEffectiveGenesis()),
	}, nil
}

type StorageSize int64

func (s StorageSize) String() string {
	return units.BytesSize(float64(s))
}

type OverViewStatus struct {
	State          pb.PostSetupStatus_State
	CompletedSize  StorageSize
	CommitmentSize StorageSize

	CoinBase  string
	SmesherId []byte
	GenesisId types.Hash20

	IsSynced        bool
	CurrentLayerId  int
	GenesisEndLayer int
}

func resolveToLocalhost(URL string) string {
	return strings.ReplaceAll(URL, "0.0.0.0", "127.0.0.1")
}

func init() {
	rootCmd.AddCommand(overviewCmd)
}
