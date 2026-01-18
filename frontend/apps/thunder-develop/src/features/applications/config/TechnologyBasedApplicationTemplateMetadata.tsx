/**
 * Copyright (c) 2025, WSO2 LLC. (https://www.wso2.com).
 *
 * WSO2 LLC. licenses this file to you under the Apache License,
 * Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

import type {ApplicationTemplate, ApplicationTemplateMetadata} from '../models/application-templates';
import {TechnologyApplicationTemplate} from '../models/application-templates';
import ReactTemplate from '../data/application-templates/technology-based/react.json';
import NextJSTemplate from '../data/application-templates/technology-based/nextjs.json';

const TechnologyBasedApplicationTemplateMetadata: ApplicationTemplateMetadata<TechnologyApplicationTemplate>[] = [
  {
    value: TechnologyApplicationTemplate.REACT,
    icon: (
      <svg xmlns="http://www.w3.org/2000/svg" width="40" height="40" viewBox="0 0 16 16" fill="#149ECA">
        <path
          fillRule="evenodd"
          clipRule="evenodd"
          d="M2.769 5.92C1.414 6.53.762 7.295.762 7.999c0 .703.652 1.467 2.007 2.077 1.319.593 3.168.97 5.231.97 2.063 0 3.912-.377 5.231-.97 1.355-.61 2.007-1.374 2.007-2.077 0-.704-.652-1.468-2.007-2.077C11.912 5.327 10.063 4.95 8 4.95c-2.063 0-3.912.377-5.231.97Zm-.313-.694C3.895 4.579 5.855 4.188 8 4.188c2.145 0 4.105.39 5.544 1.038C14.946 5.857 16 6.808 16 7.998c0 1.19-1.054 2.14-2.456 2.771-1.439.648-3.399 1.038-5.544 1.038-2.145 0-4.105-.39-5.544-1.038C1.054 10.14 0 9.188 0 7.998s1.054-2.141 2.456-2.772Z"
        />
        <path
          fillRule="evenodd"
          clipRule="evenodd"
          d="M7.183 2.429c-1.205-.869-2.193-1.052-2.802-.7-.61.352-.945 1.298-.795 2.777.145 1.439.743 3.229 1.775 5.015 1.031 1.787 2.282 3.2 3.456 4.045 1.205.869 2.193 1.052 2.802.7.61-.352.945-1.298.795-2.777-.145-1.439-.743-3.229-1.775-5.015-1.031-1.787-2.282-3.2-3.456-4.045Zm.445-.618c1.28.922 2.598 2.424 3.671 4.282 1.073 1.857 1.715 3.75 1.873 5.32.155 1.53-.142 2.918-1.172 3.513-1.03.595-2.38.158-3.629-.741-1.28-.923-2.598-2.425-3.67-4.283-1.073-1.857-1.715-3.75-1.873-5.32C2.673 3.052 2.969 1.664 4 1.07c1.03-.595 2.38-.157 3.628.742Z"
        />
        <path
          fillRule="evenodd"
          clipRule="evenodd"
          d="M12.414 4.506c.15-1.478-.186-2.425-.795-2.777-.61-.352-1.597-.169-2.802.7-1.174.845-2.425 2.258-3.456 4.045-1.032 1.786-1.63 3.576-1.775 5.015-.15 1.479.186 2.425.795 2.777.61.352 1.597.169 2.802-.7 1.174-.845 2.425-2.258 3.456-4.045 1.032-1.786 1.63-3.576 1.775-5.015Zm.758.076c-.158 1.57-.8 3.463-1.873 5.32-1.072 1.858-2.39 3.36-3.67 4.283-1.248.899-2.598 1.336-3.629.74-1.03-.594-1.327-1.982-1.172-3.512.158-1.57.8-3.462 1.873-5.32 1.072-1.858 2.39-3.36 3.67-4.282C9.62.91 10.97.474 12 1.069c1.03.595 1.327 1.983 1.172 3.513Z"
        />
        <path d="M8 9.521a1.524 1.524 0 1 0 0-3.047A1.524 1.524 0 0 0 8 9.52Z" />
      </svg>
    ),
    titleKey: 'applications:onboarding.configure.stack.technology.react.title',
    descriptionKey: 'applications:onboarding.configure.stack.technology.react.description',
    template: ReactTemplate as ApplicationTemplate,
  },
  {
    value: TechnologyApplicationTemplate.NEXTJS,
    icon: (
      <svg xmlns="http://www.w3.org/2000/svg" width="40" height="40" viewBox="0 0 16 16" fill="none">
        <path d="M8 15.733A7.733 7.733 0 1 0 8 .267a7.733 7.733 0 0 0 0 15.466Z" fill="#000" />
        <path
          fillRule="evenodd"
          clipRule="evenodd"
          d="M8 .533a7.467 7.467 0 1 0 0 14.934A7.467 7.467 0 0 0 8 .533ZM0 8a8 8 0 1 1 16 0A8 8 0 0 1 0 8Z"
          fill="#fff"
        />
        <path
          d="M13.29 14.002 6.146 4.8H4.8v6.397h1.077v-5.03l6.567 8.486c.297-.198.58-.416.846-.651Z"
          fill="url(#b)"
        />
        <path d="M11.289 4.8h-1.067v6.4h1.067V4.8Z" fill="url(#c)" />
        <defs>
          <linearGradient id="b" x1="9.689" y1="10.355" x2="12.845" y2="14.267" gradientUnits="userSpaceOnUse">
            <stop stopColor="#fff" />
            <stop offset="1" stopColor="#fff" stopOpacity="0" />
          </linearGradient>
          <linearGradient id="c" x1="10.755" y1="4.8" x2="10.738" y2="9.5" gradientUnits="userSpaceOnUse">
            <stop stopColor="#fff" />
            <stop offset="1" stopColor="#fff" stopOpacity="0" />
          </linearGradient>
        </defs>
      </svg>
    ),
    titleKey: 'applications:onboarding.configure.stack.technology.nextjs.title',
    descriptionKey: 'applications:onboarding.configure.stack.technology.nextjs.description',
    template: NextJSTemplate as ApplicationTemplate,
    disabled: false,
  },
  {
    value: TechnologyApplicationTemplate.OTHER,
    icon: (
      <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 16 16" fill="none">
        <path
          fillRule="evenodd"
          clipRule="evenodd"
          d="M8.658 7.036c.207-.053.386-.1.563-.148l.01-.002c.218-.058.433-.115.694-.182.09-.022.12-.06.075-.128a2.364 2.364 0 0 0-.875-1.063c-.61-.415-1.29-.498-1.998-.37a3.347 3.347 0 0 0-2.112 1.305c-.498.649-.739 1.388-.641 2.21.083.693.4 1.252.958 1.674.603.452 1.29.596 2.028.498.898-.121 1.607-.566 2.15-1.282.123-.163.228-.333.317-.511.127.459.382.853.754 1.174.528.453 1.154.649 1.84.664.067-.008.135-.014.203-.02h.002c.131-.012.265-.023.398-.048a3.456 3.456 0 0 0 1.81-.965c.701-.71.995-1.539.867-2.557-.098-.724-.467-1.274-1.063-1.682-.656-.445-1.38-.52-2.15-.384-.897.158-1.56.535-2.11 1.26a3.203 3.203 0 0 0-.523.98H7.428a.204.204 0 0 0-.188.12 12.03 12.03 0 0 0-.385.822c-.053.128-.015.226.143.226h1.456l-.054.077-.002.002v.001a3.434 3.434 0 0 1-.148.2c-.339.384-.769.565-1.282.497-.595-.083-1.01-.58-1.018-1.184a1.604 1.604 0 0 1 .77-1.433c.43-.279.89-.347 1.364-.113.159.076.242.159.355.287.094.109.105.106.206.079l.013-.004Zm5.553.714v-.003c-.004-.061-.007-.114-.018-.168-.136-.746-.822-1.169-1.538-1.003-.702.159-1.154.604-1.32 1.312-.136.589.15 1.184.694 1.426.414.18.83.158 1.229-.045.596-.31.92-.792.958-1.44a4.066 4.066 0 0 1-.005-.079Z"
          fill="#00ACD7"
        />
        <path
          d="M1.388 6.802c-.015.023-.008.038.022.038l2.73.008c.023 0 .06-.023.076-.046l.128-.196c.015-.022.007-.045-.023-.045H1.63c-.03 0-.068.015-.083.038l-.158.203ZM.271 7.534c-.03 0-.037-.015-.022-.038l.158-.203c.015-.023.053-.038.083-.038h3.44c.03 0 .044.023.037.045l-.06.181c-.008.03-.038.045-.068.045L.27 7.534ZM2.097 8.228c-.03 0-.038-.023-.023-.045l.106-.189c.015-.023.045-.045.075-.045h1.508c.03 0 .045.022.045.053l-.015.18c0 .03-.03.053-.052.053l-1.644-.007Z"
          fill="#00ACD7"
        />
      </svg>
    ),
    titleKey: 'Go',
    descriptionKey: 'applications:onboarding.configure.stack.technology.go.description',
    template: null as unknown as ApplicationTemplate,
    disabled: false,
  },
  {
    value: TechnologyApplicationTemplate.OTHER,
    icon: (
      <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 16 16" fill="none">
        <path d="M6.033 9.48h3.95L8.008 4.677 6.033 9.48Z" fill="url(#a)" />
        <path d="M6.033 9.48h3.95L8.008 4.677 6.033 9.48Z" fill="url(#b)" />
        <path d="m15.235 3.149-.519 8.195L9.724.6l5.511 2.549Z" fill="url(#c)" />
        <path d="m15.235 3.149-.519 8.195L9.724.6l5.511 2.549Z" fill="url(#d)" />
        <path d="M6.292.6.782 3.15l.518 8.194L6.292.6Z" fill="url(#e)" />
        <path d="M6.292.6.782 3.15l.518 8.194L6.292.6Z" fill="url(#f)" />
        <path d="M11.012 11.89H5.005l-.768 1.858 3.771 2.15 3.772-2.15-.768-1.858Z" fill="url(#g)" />
        <path d="M11.012 11.89H5.005l-.768 1.858 3.771 2.15 3.772-2.15-.768-1.858Z" fill="url(#h)" />
        <defs>
          <linearGradient id="a" x1="11.25" y1="5.938" x2="6.937" y2="9.25" gradientUnits="userSpaceOnUse">
            <stop stopColor="#AD23F4" />
            <stop offset="1" stopColor="#EA1A85" />
          </linearGradient>
          <linearGradient id="b" x1="11.25" y1="5.938" x2="6.937" y2="9.25" gradientUnits="userSpaceOnUse">
            <stop stopColor="#AD23F4" />
            <stop offset="1" stopColor="#EA1A85" />
          </linearGradient>
          <linearGradient id="c" x1="9.75" y1=".625" x2="14.625" y2="11.375" gradientUnits="userSpaceOnUse">
            <stop stopColor="#D22FF2" />
            <stop offset="1" stopColor="#8A18F4" />
          </linearGradient>
          <linearGradient id="d" x1="9.75" y1=".625" x2="14.625" y2="11.375" gradientUnits="userSpaceOnUse">
            <stop stopColor="#D22FF2" />
            <stop offset="1" stopColor="#8A18F4" />
          </linearGradient>
          <linearGradient id="e" x1="5.563" y1="-.125" x2="1.188" y2="11.375" gradientUnits="userSpaceOnUse">
            <stop stopColor="#EC34E7" />
            <stop offset="1" stopColor="#F1206E" />
          </linearGradient>
          <linearGradient id="f" x1="5.563" y1="-.125" x2="1.188" y2="11.375" gradientUnits="userSpaceOnUse">
            <stop stopColor="#EC34E7" />
            <stop offset="1" stopColor="#F1206E" />
          </linearGradient>
          <linearGradient id="g" x1="11.75" y1="12.625" x2="8" y2="13.875" gradientUnits="userSpaceOnUse">
            <stop stopColor="#C2099F" />
            <stop offset="1" stopColor="#F40F4A" />
          </linearGradient>
          <linearGradient id="h" x1="11.75" y1="12.625" x2="8" y2="13.875" gradientUnits="userSpaceOnUse">
            <stop stopColor="#C2099F" />
            <stop offset="1" stopColor="#F40F4A" />
          </linearGradient>
        </defs>
      </svg>
    ),
    titleKey: 'Angular',
    descriptionKey: 'applications:onboarding.configure.stack.technology.angular.description',
    template: null as unknown as ApplicationTemplate,
    disabled: false,
  },
  {
    value: TechnologyApplicationTemplate.OTHER,
    icon: (
      <svg xmlns="http://www.w3.org/2000/svg" preserveAspectRatio="xMidYMid" viewBox="-45 0 346 346">
        <path
          fill="#5382A1"
          d="M82.554 267.473s-13.198 7.675 9.393 10.272c27.369 3.122 41.356 2.675 71.517-3.034 0 0 7.93 4.972 19.003 9.279-67.611 28.977-153.019-1.679-99.913-16.517M74.292 229.659s-14.803 10.958 7.805 13.296c29.236 3.016 52.324 3.263 92.276-4.43 0 0 5.526 5.602 14.215 8.666-81.747 23.904-172.798 1.885-114.296-17.532"
        />
        <path
          fill="#E76F00"
          d="M143.942 165.515c16.66 19.18-4.377 36.44-4.377 36.44s42.301-21.837 22.874-49.183c-18.144-25.5-32.059-38.172 43.268-81.858 0 0-118.238 29.53-61.765 94.6"
        />
        <path
          fill="#5382A1"
          d="M233.364 295.442s9.767 8.047-10.757 14.273c-39.026 11.823-162.432 15.393-196.714.471-12.323-5.36 10.787-12.8 18.056-14.362 7.581-1.644 11.914-1.337 11.914-1.337-13.705-9.655-88.583 18.957-38.034 27.15 137.853 22.356 251.292-10.066 215.535-26.195M88.9 190.48s-62.771 14.91-22.228 20.323c17.118 2.292 51.243 1.774 83.03-.89 25.978-2.19 52.063-6.85 52.063-6.85s-9.16 3.923-15.787 8.448c-63.744 16.765-186.886 8.966-151.435-8.183 29.981-14.492 54.358-12.848 54.358-12.848M201.506 253.422c64.8-33.672 34.839-66.03 13.927-61.67-5.126 1.066-7.411 1.99-7.411 1.99s1.903-2.98 5.537-4.27c41.37-14.545 73.187 42.897-13.355 65.647 0 .001 1.003-.895 1.302-1.697"
        />
        <path
          fill="#E76F00"
          d="M162.439.371s35.887 35.9-34.037 91.101c-56.071 44.282-12.786 69.53-.023 98.377-32.73-29.53-56.75-55.526-40.635-79.72C111.395 74.612 176.918 57.393 162.439.37"
        />
        <path
          fill="#5382A1"
          d="M95.268 344.665c62.199 3.982 157.712-2.209 159.974-31.64 0 0-4.348 11.158-51.404 20.018-53.088 9.99-118.564 8.824-157.399 2.421.001 0 7.95 6.58 48.83 9.201"
        />
      </svg>
    ),
    titleKey: 'Java',
    descriptionKey: 'applications:onboarding.configure.stack.technology.java.description',
    template: null as unknown as ApplicationTemplate,
    disabled: false,
  },
  {
    value: TechnologyApplicationTemplate.OTHER,
    icon: (
      <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 16 16" fill="none">
        <path
          d="M8.001.832a9.959 9.959 0 0 0-1.666.143c-1.476.26-1.744.806-1.744 1.813v1.329h3.488v.443H3.28c-1.013 0-1.9.61-2.178 1.768-.32 1.329-.334 2.158 0 3.545.248 1.033.84 1.768 1.854 1.768h1.199v-1.593c0-1.151.996-2.167 2.179-2.167h3.483c.97 0 1.744-.798 1.744-1.772V2.788c0-.946-.797-1.656-1.744-1.813C9.22.875 8.598.83 8.001.832Zm-1.886 1.07c.36 0 .655.298.655.666a.66.66 0 0 1-.655.663.658.658 0 0 1-.654-.663.66.66 0 0 1 .654-.667Z"
          fill="url(#python-a)"
        />
        <path
          d="M11.997 4.56v1.549c0 1.2-1.018 2.211-2.179 2.211H6.335c-.955 0-1.744.817-1.744 1.772v3.322c0 .945.822 1.5 1.744 1.772 1.104.325 2.162.383 3.483 0 .878-.254 1.744-.766 1.744-1.772v-1.33H8.08v-.443h5.227c1.014 0 1.392-.707 1.744-1.768.364-1.093.349-2.143 0-3.545-.25-1.009-.729-1.768-1.744-1.768h-1.309Zm-1.96 8.41c.362 0 .655.297.655.663a.66.66 0 0 1-.654.667.662.662 0 0 1-.655-.667.66.66 0 0 1 .655-.663Z"
          fill="url(#python-b)"
        />
        <defs>
          <linearGradient id="python-a" x1=".857" y1=".832" x2="8.9" y2="7.685" gradientUnits="userSpaceOnUse">
            <stop stopColor="#5A9FD4" />
            <stop offset="1" stopColor="#306998" />
          </linearGradient>
          <linearGradient id="python-b" x1="9.954" y1="13.521" x2="7.069" y2="9.478" gradientUnits="userSpaceOnUse">
            <stop stopColor="#FFD43B" />
            <stop offset="1" stopColor="#FFE873" />
          </linearGradient>
        </defs>
      </svg>
    ),
    titleKey: 'Python',
    descriptionKey: 'applications:onboarding.configure.stack.technology.python.description',
    template: null as unknown as ApplicationTemplate,
    disabled: false,
  },
  {
    value: TechnologyApplicationTemplate.OTHER,
    icon: (
      <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 456 456" fill="none">
        <rect width="456" height="456" fill="#512BD4" />
        <path
          d="M81.2738 291.333C78.0496 291.333 75.309 290.259 73.052 288.11C70.795 285.906 69.6665 283.289 69.6665 280.259C69.6665 277.173 70.795 274.529 73.052 272.325C75.309 270.121 78.0496 269.019 81.2738 269.019C84.5518 269.019 87.3193 270.121 89.5763 272.325C91.887 274.529 93.0424 277.173 93.0424 280.259C93.0424 283.289 91.887 285.906 89.5763 288.11C87.3193 290.259 84.5518 291.333 81.2738 291.333Z"
          fill="white"
        />
        <path
          d="M210.167 289.515H189.209L133.994 202.406C132.597 200.202 131.441 197.915 130.528 195.546H130.044C130.474 198.081 130.689 203.508 130.689 211.827V289.515H112.149V171H134.477L187.839 256.043C190.096 259.57 191.547 261.994 192.192 263.316H192.514C191.977 260.176 191.708 254.859 191.708 247.365V171H210.167V289.515Z"
          fill="white"
        />
        <path
          d="M300.449 289.515H235.561V171H297.87V187.695H254.746V221.249H294.485V237.861H254.746V272.903H300.449V289.515Z"
          fill="white"
        />
        <path
          d="M392.667 187.695H359.457V289.515H340.272V187.695H307.143V171H392.667V187.695Z"
          fill="white"
        />
      </svg>
    ),
    titleKey: '.NET',
    descriptionKey: 'applications:onboarding.configure.stack.technology.dotnet.description',
    template: null as unknown as ApplicationTemplate,
    disabled: false,
  },
];

export default TechnologyBasedApplicationTemplateMetadata;
