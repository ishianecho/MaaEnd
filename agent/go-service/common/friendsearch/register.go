package friendsearch

import maa "github.com/MaaXYZ/maa-framework-go/v4"

func Register() {
	maa.AgentServerRegisterCustomAction("FriendSearchInitAction", &InitAction{})
	maa.AgentServerRegisterCustomAction("FriendSearchInputAction", &InputAction{})
	maa.AgentServerRegisterCustomRecognition("FriendSearchExhaustedRecognition", &ExhaustedRecognition{})
}
