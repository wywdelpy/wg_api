package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func ReadWGCreds() error {
	var inter InterfaceGorm
	inter.Name = "wg0"
	inter.PublicKey = readfile(PubKeyPath)
	inter.PrivateKey = readfile(PrivateKeyPath)
	inter.PrivateKey = strings.TrimSpace(inter.PrivateKey)
	inter.PublicKey = strings.TrimSpace(inter.PublicKey)
	lg.Println("/etc/wireguard/ was read successfully!")
	err := AddInterfaceToORM(inter)
	if err != nil {
		err = fmt.Errorf("Failed to add iterface to orm: %s", err)
		return err
	}
	return nil
}

func createClientConfig(inter InterfaceGorm, peer PeerGorm) WgConfig {
	var cfg WgConfig
	cfg.FileName = peer.Name
	cfg.Interface.Address = peer.AllowedIP
	cfg.Interface.PrivateKey = peer.PrivateKey
	cfg.Interface.MTU = MTU
	cfg.Interface.DNS = DNS
	cfg.Peer.PublicKey = inter.PublicKey
	cfg.Peer.AllowedIPs = "0.0.0.0/0"
	cfg.Peer.Endpoint = WG_ENDPOINT
	cfg.Peer.PersistentKeepalive = "21"
	return cfg
}

func setPeers(peers []PeerGorm) error {
	for _, peer := range peers {
		cmd := exec.Command("wg", "set", "wg0", "peer", peer.PublicKey, "allowed-ips", peer.AllowedIP)
		// Запускаем команду и возвращаем ошибку, если она произошла
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("error executing command set: %v", err)
		}
		lg.Println(peer.AllowedIP)
		lg.Println(peer.PublicKey)
	}
	cmd := exec.Command("wg-quick", "save", "wg0")
	// Запускаем команду и возвращаем ошибку, если она произошла
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error executing command save: %v", err)
	}
	return nil
}

func setPeer(peer PeerGorm) error {
	cmd := exec.Command("wg", "set", "wg0", "peer", peer.PublicKey, "allowed-ips", peer.AllowedIP)
	// Запускаем команду и возвращаем ошибку, если она произошла
	if err := cmd.Run(); err != nil {
		lgError.Printf("error executing command set: %v", err)
		return fmt.Errorf("error executing command set: %v", err)
	}
	cmd = exec.Command("wg-quick", "save", "wg0")
	// Запускаем команду и возвращаем ошибку, если она произошла
	if err := cmd.Run(); err != nil {
		lgError.Printf("error executing command save: %v", err)
		return fmt.Errorf("error executing command save: %v", err)
	}
	lgWG.Printf("Peer %s was set allowed_ip %s", peer.Name, peer.AllowedIP)
	return nil
}

func writePeersIntoWgConf(filePath string, peers []PeerGorm) error {
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		lg.Printf("Failed to open %s:%s", filePath, err)
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	for _, peer := range peers {
		str := fmt.Sprintf("[Peer]\nPublicKey = %s\nAllowedIPs = %s\n\n", peer.PublicKey, peer.AllowedIP)
		_, err := writer.Write([]byte(str))
		if err != nil {
			lg.Printf("Failed to writed data to %s:%s", filePath, err)
			return err
		}
	}
	return nil
}

func grantConsumerPeer(cons ConsGorm, month, days int) (ConsGorm, PeerGorm, error) {
	var vacantPeer PeerGorm
	vacantPeer, err := GetVacantPeerFromORM(month, days)
	if err != nil {
		lgError.Printf("Failed to get vacant peer from database: %s", err)
		return ConsGorm{}, PeerGorm{}, fmt.Errorf("Failed to get vacant peer from database: %s", err)
	}
	var resCons ConsGorm
	var resPeer PeerGorm
	resCons, resPeer, err = grantConsumerPeerInORM(cons, vacantPeer)
	if err != nil {
		lgError.Printf("Failed to grant peer to consumer in database: %s", err)
		err = fmt.Errorf("Failed to grant peer to consumer in database: %s", err)
		return ConsGorm{}, PeerGorm{}, err
	}
	lgWG.Printf("User @%s was granted peer: %s expiration_time %s allowed_ip %s successfully!", resCons.Username, resPeer.Name, resPeer.ExpirationTime, resPeer.AllowedIP)
	return resCons, resPeer, nil

}

func GenAndWritePeers() error {
	peers := generatePeers()
	err := setPeers(peers)
	if err != nil {
		lg.Printf("Failed to write peers into wg conf: %s", err)
		return err
	}
	err = writePeersToORM(peers)
	if err != nil {
		lg.Printf("Failed to write peers into ORM: %s", err)
		return err
	}
	return nil
}

func GiveLastPaidPeer(cons ConsGorm) (ConsGorm, PeerGorm, error) {
	var resCons ConsGorm
	var resPeer PeerGorm
	resCons, resPeer, err := GiveLastPaidPeerFromORM(cons)
	if err != nil {
		err = fmt.Errorf("Failed to give last paid peer of user %s from ORM : %s", resCons.Username, err)
		return ConsGorm{}, PeerGorm{}, err
	}
	return resCons, resPeer, nil
}

func RestrictPeer(peer PeerGorm) error {
	peer.AllowedIP = turnOffPeer(peer.AllowedIP)
	peer.Status = "Expired"
	cmd := exec.Command("wg", "set", "wg0", "peer", peer.PublicKey, "allowed-ips", peer.AllowedIP)
	// Запускаем команду и возвращаем ошибку, если она произошла
	if err := cmd.Run(); err != nil {
		lgError.Printf("error executing command set: %v", err)
		return fmt.Errorf("error executing command set: %v", err)
	}
	cmd = exec.Command("wg-quick", "save", "wg0")
	// Запускаем команду и возвращаем ошибку, если она произошла
	if err := cmd.Run(); err != nil {
		lgError.Printf("error executing command save: %v", err)
		return fmt.Errorf("error executing command save: %v", err)
	}
	if err := RestictPeerInORM(peer); err != nil {
		lgError.Printf("Failed to restrict peer %d in ORM: %s", peer.ID, err)
		return fmt.Errorf("Failed to restrict peer %d in ORM: %s", peer.ID, err)
	}
	lgError.Printf("Peer %s was restircted succefully. Allowed_ip: %s", peer.PublicKey, peer.AllowedIP)
	return nil
}

func KillAndRegenPeer(oldPeer PeerGorm) (ConsGorm, error) {
	// Регенерируем пира в ORM и получаем обновлённого пира и удалённую ассоциацию
	newPeer, oldCons, err := KillAndRegenPeerInORM(oldPeer)
	if err != nil {
		return ConsGorm{}, fmt.Errorf("failed to kill and regen peer in ORM %s: %v", oldPeer.PublicKey, err)
	}

	// Устанавливаем нового пира в WireGuard (wg set wg0 peer <newPeer.PublicKey> allowed-ips <newPeer.AllowedIP>)
	cmd := exec.Command("wg", "set", "wg0", "peer", newPeer.PublicKey, "allowed-ips", newPeer.AllowedIP)
	if err := cmd.Run(); err != nil {
		return ConsGorm{}, fmt.Errorf("error executing command set: %v", err)
	}

	// Удаляем старого пира из WireGuard (wg set wg0 peer <oldPeer.PublicKey> remove)
	cmd = exec.Command("wg", "set", "wg0", "peer", oldPeer.PublicKey, "remove")
	if err := cmd.Run(); err != nil {
		return ConsGorm{}, fmt.Errorf("error executing command remove: %v", err)
	}

	// Сохраняем изменения в конфигурации WireGuard (wg-quick save wg0)
	cmd = exec.Command("wg-quick", "save", "wg0")
	if err := cmd.Run(); err != nil {
		return ConsGorm{}, fmt.Errorf("error executing command save: %v", err)
	}

	lgWG.Printf("Peer %s was killed and regened to %s", oldPeer.PublicKey, newPeer.PublicKey)
	return oldCons, nil
}
