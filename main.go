package main

import (
	"fmt"
	"log"
	"os"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)


var(
	vaultDir="~/.gotion"
	cursorStyle= lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	docStyle=lipgloss.NewStyle().Margin(1,2)
)

type item struct{
	title,desc string
}

func (i item) Title() string {
	return i.title
}

func (i item) Description() string {
	return i.desc
}
func (i item) FilterValue() string {
	return i.title
}

type model struct {
	newFileInput           textinput.Model
	createFileInputVisible bool
	currentFile *os.File
	noteTextArea textarea.Model
	list list.Model
	showList bool
	
}

func init(){
	homeDir,err:=os.UserHomeDir()
	if err!=nil{
		log.Fatal("Error getting home directory",err)
	}
	vaultDir=fmt.Sprintf("%s/.gotion",homeDir)
}

func initialModel() model {
	err:=os.MkdirAll(vaultDir, 0755)
	if err!=nil{
		log.Fatal(err)
	}

	ti := textinput.New()
	ti.Placeholder = "Enter file name"
	ti.Focus()
	ti.CharLimit = 156
	ti.Width=30
	ti.Cursor.Style=cursorStyle
	ti.PromptStyle=cursorStyle


	//text area
	ta:=textarea.New()
	ta.Placeholder="Write your note here..."
	ta.Focus()

	//list 

	noteList:=listFiles()
	finalList:=list.New(noteList,list.NewDefaultDelegate(),0,0)
	finalList.Title="All Notes"
	finalList.Styles.Title=lipgloss.NewStyle().Foreground(lipgloss.Color("16")).Background(lipgloss.Color("254")).Padding(0,1)
	return model{
		newFileInput:           ti,
		createFileInputVisible: false,
		noteTextArea: ta,
		list:  	finalList,
		showList: false,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "ctrl+n":
			m.createFileInputVisible = true
			return m, nil
		case "enter":
			if m.createFileInputVisible{
				m.createFileInputVisible = false
				filename:=m.newFileInput.Value()
				if filename!=""{
					filePath:=fmt.Sprintf("%s/%s.md",vaultDir,filename)
					if _,err:=os.Stat(filePath);err==nil{
						return m,nil
					}

					f,err:=os.Create(filePath)
					if err!=nil{
						log.Fatalf("%v",err)
					}
					m.currentFile=f
					
					if err!=nil{
						log.Fatal(err)
					}
					m.list.SetItems(listFiles())
				m.newFileInput.SetValue("")

				}
				return m, nil
			}
			if m.showList {
				selectedItem,ok:=m.list.SelectedItem().(item)
				if ok{
					filepath:=fmt.Sprintf("%s/%s",vaultDir,selectedItem.title)
					content,err:=os.ReadFile(filepath) 
					if err!=nil{
						log.Printf("Error reading file %s",filepath)
						return m,nil
					}
					m.noteTextArea.SetValue(string(content))

					m.currentFile,err=os.OpenFile(filepath,os.O_RDWR,0644)
					m.showList=false
					if err!=nil{
						log.Printf("Error opening file %s",filepath)
						return m,nil
					}
				}
				return  m,nil

			}
			
		case "ctrl+s":
			if m.currentFile==nil{
				break
			}
			if err:=m.currentFile.Truncate(0);err!=nil{
				fmt.Println("Can not save the file")
				return m,nil
			} 
			if _,err:=m.currentFile.Seek(0,0); err!=nil{
				fmt.Println("Can not save the file")
				return m,nil
			}
			if _,err:=m.currentFile.WriteString(m.noteTextArea.Value()); err!=nil{
				fmt.Println("Can not save the file")
				return m,nil
			}
			if err:=m.currentFile.Close(); err!=nil{
				fmt.Println("Can not close the file")
			}
			m.currentFile=nil
			m.noteTextArea.SetValue("")
			return m,nil
		case "ctrl+l":
			m.showList=true
			return m,nil 

		case "ctrl+d":
			if m.showList{
				selectedItem,ok:=m.list.SelectedItem().(item)
				if ok{
					filepath:=fmt.Sprintf("%s/%s",vaultDir,selectedItem.title)
					if err:=os.Remove(filepath); err!=nil{
						log.Printf("Error removing file %s",filepath)
						return m,nil
					}
					m.list.SetItems(listFiles())
				}
				return m,nil
			}
			return  m,nil
		case "esc":
			if m.createFileInputVisible{
				m.createFileInputVisible=false
				m.newFileInput.SetValue("")
				return m,nil
			}
			if m.showList{
				if m.list.FilterState()!=list.Filtering{
					break
				}
				m.showList=false
				m.list.SetItems(listFiles())
				return m,nil
			}
			if m.currentFile!=nil{
				m.currentFile.Close()
				m.currentFile=nil
				m.noteTextArea.SetValue("")
			}
			return m,nil
		}
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v-5)
	}
	if m.createFileInputVisible{
		m.newFileInput,_ = m.newFileInput.Update(msg)
	
	}
	if m.currentFile!=nil{
		m.noteTextArea,_ = m.noteTextArea.Update(msg)
	}

	if m.showList{
		m.list,_=m.list.Update(msg)
		
	}
	return m, nil
}

func (m model) View() string {
	var style = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Padding(0, 2, 0, 2)

	welcome := style.Render("Welcome to Gotion")

	var helpStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#888888"))
	help := helpStyle.Render("Ctrl+N: new file ⋅ Ctrl+L: list files ⋅ Esc: back/save ⋅ Ctrl+S: save file ⋅ Ctrl+Q: quit")
	view := "🦚"
	if m.createFileInputVisible {
		view = m.newFileInput.View()
	}

	if m.currentFile!=nil{
		view=m.noteTextArea.View()
	}
	if m.showList{
		view=m.list.View()
	}
	
	return fmt.Sprintf("\n%s\n\n%s\n\n%s", welcome, view, help)
}

func main() {
	fmt.Println("Welcome to Gotion app")
	p := tea.NewProgram(initialModel())

	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}


func listFiles()[]list.Item{
	items:=make([]list.Item,0)
	entries,err:=os.ReadDir(vaultDir)
	if err!=nil{
		log.Fatal("Error reading directory",err)
	}
	for _,entry:=range entries{

		if !entry.IsDir(){
			fileInfo,err:=entry.Info()
			if err!=nil{
				continue
			}
			items=append(items,item{
				title:entry.Name(),
				desc:fileInfo.ModTime().Format("2006-01-02 15:04:05"),
			})
		}
	}
	return  items
}