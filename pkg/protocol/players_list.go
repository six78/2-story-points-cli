package protocol

type PlayersList []Player

func (l PlayersList) Get(id PlayerID) (Player, bool) {
	for _, player := range l {
		if player.ID == id {
			return player, true
		}
	}
	return Player{}, false
}
