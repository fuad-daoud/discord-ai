GREEN="\e[1;32m"
RESET="\e[0m"
PREFIX="${GREEN}======|"
echo "${PREFIX}TASK[1] INSTALLING ZSH AND PLUGINS${RESET}"

apt update
apt install zsh --yes
chsh -s $(which zsh)
sh -c "$(curl -fsSL https://raw.githubusercontent.com/ohmyzsh/ohmyzsh/master/tools/install.sh)"
apt install nodejs --yes
apt install npm --yes
apt install tree --yes


echo "${PREFIX}ADDING zsh-history-enquirer${RESET}"
npm i -g zsh-history-enquirer
echo "${PREFIX}ADDING fzf${RESET}"
apt install fzf --yes

echo "${PREFIX}ADDING zsh-autosuggestions${RESET}"
git clone https://github.com/zsh-users/zsh-autosuggestions ${ZSH_CUSTOM:-~/.oh-my-zsh/custom}/plugins/zsh-autosuggestions
echo "source ${ZSH_CUSTOM:-~/.oh-my-zsh/custom}/plugins/zsh-autosuggestions/zsh-syntax-highlighting.zsh" >> ${ZDOTDIR:-$HOME}/.zshrc

echo "${PREFIX}ADDING zsh-autosuggestions${RESET}"
git clone https://github.com/zsh-users/zsh-syntax-highlighting.git ${ZSH_CUSTOM:-~/.oh-my-zsh/custom}/plugins/zsh-syntax-highlighting
echo "source ${ZSH_CUSTOM:-~/.oh-my-zsh/custom}/plugins/zsh-syntax-highlighting/zsh-syntax-highlighting.zsh" >> ${ZDOTDIR:-$HOME}/.zshrc

source ~/.zshrc

echo "${PREFIX}FINISHED TASK[1]${RESET}"
echo "${PREFIX}TASK[2] INSTALLING GOLANG${RESET}"

curl -OL https://golang.org/dl/go1.22.4.linux-amd64.tar.gz
sha256sum go1.22.4.linux-amd64.tar.gz
tar -C /usr/local -xvf go1.22.4.linux-amd64.tar.gz
echo 'PATH=$PATH:/usr/local/go/bin' >> ~/.zshrc
echo "${PREFIX}INSTALLED GOLANG${RESET}"


echo "${PREFIX}INSTALLING FFMPEG${RESET}"

apt install ffmpeg --yes

source ~/.zshrc
go version
ffmpeg -version


echo "${PREFIX}FINISHED TASK[2]${RESET}"


echo "${PREFIX}TASK[3] SETTING UP PM2${RESET}"

npm install pm2 -g

echo "${PREFIX}FINISHED TASK[3]${RESET}"


echo "${PREFIX}FINISHED ALL TASKS${RESET}"



#use to ssh sshpass -p 1234Test ssh root@165.22.92.220

#https://medium.com/swlh/how-to-deploy-your-application-to-digital-ocean-using-github-actions-and-save-up-on-ci-cd-costs-74b7315facc2

#https://pm2.keymetrics.io/docs/usage/expose/

#https://pm2.io/docs/enterprise/collector/go/